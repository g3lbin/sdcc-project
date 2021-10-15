package rpcpeer

import (
	"context"
	"fmt"
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Peer struct {
	Ip 			string
	Algorithm	string
	Membership 	[]string
	TimeStruct 	*utils.Time
	MembersId 	map[string]int
}
// strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
var datastoreHandler struct {
	ctx context.Context
	coll *mongo.Collection
}
var queue []utils.Sender
var ackForMessages map[string]int
var messagesPerPeer map[string]int
var queueLock sync.Mutex
var expected uint64 = 0
var last uint64 = 0
var ChFromPeers chan utils.Sender
var ChAck chan utils.Sender

func InitRpcPeer(hostname string, membership []string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	datastoreHandler.ctx = context.Background()
	err = client.Connect(datastoreHandler.ctx)
	if err != nil {
		log.Fatal(err)
	}
	datastoreHandler.coll = client.Database(hostname).Collection("chat")

	ackForMessages = make(map[string]int)
	messagesPerPeer = make(map[string]int)
	for _, member := range membership {
		messagesPerPeer[member] = 0
	}
}

func EnqueueMsg(algo string, new utils.Sender, args ...int) {
	queueLock.Lock()
	if algo == "tot-ordered-decentr" {
		done := false
		membersNum := args[0]
		increment := args[1]
		for i, msg := range queue {
			if new.Timestamp[0] < msg.Timestamp[0] ||
				(new.Timestamp[0] == msg.Timestamp[0] && new.Host < msg.Host) {
				tail := append([]utils.Sender{}, queue[i:]...)
				queue = append(queue[:i], new)
				queue = append(queue, tail...)

				done = true
				break
			}
		}
		if !done { // Append new message at the end of queue
			queue = append(queue, new)
		}

		if val, ok := messagesPerPeer[new.Host]; ok {
			messagesPerPeer[new.Host] = val + 1
		} else {
			messagesPerPeer[new.Host] = 1
		}
		submitAck(new, membersNum, increment)
	} else if algo == "causally-ordered-decentr" {
		queue = append(queue, new)
	}
	queueLock.Unlock()
}

func dequeue(algo string, index ...int) {
	if algo == "tot-ordered-decentr" {
		head := queue[0]
		messagesPerPeer[head.Host]--
		if len(queue) == 1 {
			queue = []utils.Sender{}
		} else {
			queue = queue[1:]
		}
	} else if algo == "causally-ordered-decentr" {
		queue = append(queue[:index[0]], queue[index[0]+1:]...)
	}
}

func checkMsgForEachPeer(membership []string) bool {
	for _, member := range membership {
		if messagesPerPeer[member] == 0 {
			return false
		}
	}

	return true
}

func submitAck(ack utils.Sender, membersNum int, increment int) {
	msgId := ack.Host + ":" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ack.Timestamp)), ","), "[]")
	if val, ok := ackForMessages[msgId]; ok {
		ackForMessages[msgId] = val + increment
	} else {
		ackForMessages[msgId] = increment
	}
}

func (p *Peer) ReceiveMessage(arg utils.Sender, res *int) error {
	if p.Algorithm == "tot-ordered-centr" {
		_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, arg)
		if err != nil {
			utils.ErrorHandler("InsertOne", err)
		}
		for ok := true; ok; {
			var res utils.Sender
			err = datastoreHandler.coll.FindOne(datastoreHandler.ctx, bson.D{{"_id", expected}}).Decode(&res)
			if err != nil {
				// ErrNoDocuments means that the filter did not match any documents in
				// the collection.
				if err == mongo.ErrNoDocuments {
					ok = false
				} else {
					utils.ErrorHandler("FindOne", err)
				}
			} else {
				ChFromPeers <- res
				expected++
			}
		}
	} else if p.Algorithm == "tot-ordered-decentr" {
		var ack utils.Sender
		var headOfQueue utils.Sender

		if arg.Type == "update" {
			// Update logical clock for 'receive' event
			p.TimeStruct.Lock.Lock()
			if p.TimeStruct.Clock[0] < arg.Timestamp[0] {
				p.TimeStruct.Clock[0] = arg.Timestamp[0]
			}
			p.TimeStruct.Clock[0]++
			p.TimeStruct.Lock.Unlock()

			EnqueueMsg(p.Algorithm, arg, len(p.Membership), 2)

			ack.Host = arg.Host
			ack.Timestamp = make([]uint64, 1)
			ack.Timestamp[0] = arg.Timestamp[0]
			ack.Type = "ack"
			ChAck <- ack
		} else {
			queueLock.Lock()
			submitAck(arg, len(p.Membership), 1)
			queueLock.Unlock()
		}
		// Check for messages to deliver to application
		for {
			queueLock.Lock()
			if len(queue) != 0 {
				headOfQueue = queue[0]
			} else {
				break
			}
			headId := headOfQueue.Host + ":" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(headOfQueue.Timestamp)), ","), "[]")
			if ackForMessages[headId] == len(p.Membership) && checkMsgForEachPeer(p.Membership) {
				dequeue(p.Algorithm)
				delete(ackForMessages, headId)

				headOfQueue.ID = expected
				expected++
				_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, headOfQueue)
				if err != nil {
					utils.ErrorHandler("InsertOne", err)
				}
			} else {
				break
			}
			queueLock.Unlock()
		}
		for last < expected {
			var delivery utils.Sender
			datastoreHandler.coll.FindOne(datastoreHandler.ctx, bson.D{{"_id", last}}).Decode(&delivery)
			ChFromPeers <- delivery
			last++
		}
		queueLock.Unlock()
	} else if p.Algorithm == "causally-ordered-decentr" {
		acceptableMsg := true
		p.TimeStruct.Lock.Lock()
		for i := 0; i < len(p.Membership); i++ {
			if (i != p.MembersId[arg.Host] && arg.Timestamp[i] > p.TimeStruct.Clock[i]) ||
				(i == p.MembersId[arg.Host] && arg.Timestamp[i] != p.TimeStruct.Clock[i] + 1) {
				acceptableMsg = false
				break
			}
		}
		p.TimeStruct.Lock.Unlock()
		if !acceptableMsg {
			EnqueueMsg(p.Algorithm, arg)
		} else {
			VectDelivery(arg, p.TimeStruct, p.MembersId, true)
		}
	}
	*res = 1

	return nil
}

func VectDelivery(new utils.Sender, time *utils.Time, membersId map[string]int, incClock bool) {
	algorithm := "causally-ordered-decentr"
	// Deliver new message
	if incClock {
		time.Lock.Lock()
		time.Clock[membersId[new.Host]]++
		time.Lock.Unlock()
	}
	queueLock.Lock()
	new.ID = expected
	expected++
	queueLock.Unlock()
	_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, new)
	if err != nil {
		utils.ErrorHandler("InsertOne", err)
	}
	ChFromPeers <- new
	// Check for others to deliver
	queueLock.Lock()
	for j, msg := range queue {
		acceptableMsg := true
		time.Lock.Lock()
		for i := 0; i < len(membersId); i++ {
			if (i != membersId[msg.Host] && msg.Timestamp[i] > time.Clock[i]) ||
				(i == membersId[msg.Host] && msg.Timestamp[i] != time.Clock[i] + 1) {
				acceptableMsg = false
				break
			}
		}
		if acceptableMsg {
			time.Clock[membersId[msg.Host]]++
		}
		time.Lock.Unlock()

		msg.ID = expected
		expected++
		_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, msg)
		if err != nil {
			utils.ErrorHandler("InsertOne", err)
		}
		dequeue(algorithm, j)
		ChFromPeers <- msg
	}
	queueLock.Unlock()
}
