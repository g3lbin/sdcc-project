package rpcpeer

import (
	"context"
	"fmt"
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Peer struct {
	Hostname	string
	Algorithm	string
	Membership 	[]string
	TimeStruct 	*utils.Time
}
// strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
var datastoreHandler struct {
	ctx context.Context
	coll *mongo.Collection
}
var queue []utils.Sender
var ackForMessages map[string]int
var queueLock sync.Mutex
var expected uint64 = 0
var last uint64 = 0
var ChFromPeers chan utils.Sender
var ChAck chan utils.Sender

func InitRpcPeer(hostname string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	datastoreHandler.ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(datastoreHandler.ctx)
	if err != nil {
		log.Fatal(err)
	}
	datastoreHandler.coll = client.Database(hostname).Collection("chat")

	ackForMessages = make(map[string]int)
}

func EnqueueMsg(new utils.Sender, membersNum int, increment int) {
	done := false
	queueLock.Lock()
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
	if !done {	// Append new message at the end of queue
		queue = append(queue, new)
	}
	submitAck(new, membersNum, increment)
	// DA RIMUOVERE
	for _, msg := range queue {
		fmt.Println(msg)
	}
	queueLock.Unlock()
}

func dequeue(msgId string) {
	//for _, msg := range queue {
	//	if msg.Host == ack.Host && msg.Timestamp[0] == ack.Timestamp[0] {
	//if msg.RcvdAck == membersNum {
	//	msg.ID = expected
	//	expected++
	//	_, err := coll.InsertOne(datastoreHandler.ctx, msg)
	//	if err != nil {
	//		utils.ErrorHandler("InsertOne", err)
	//	}
	//	queue = append(queue[:i], queue[i+1:]...)
	//}
	//	}
	//}
}

func submitAck(ack utils.Sender, membersNum int, increment int) {
	msgId := ack.Host + ":" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ack.Timestamp)), ","), "[]")
	if val, ok := ackForMessages[msgId]; ok {
		ackForMessages[msgId] = val + increment
	} else {
		ackForMessages[msgId] = increment
	}
	if ackForMessages[msgId] == membersNum {
		dequeue(msgId)
		delete(ackForMessages, msgId)
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
		var res utils.Sender
		if arg.Type == "update" {
			// Update logical clock for 'receive' event
			p.TimeStruct.Lock.Lock()
			if p.TimeStruct.Clock[0] < arg.Timestamp[0] {
				p.TimeStruct.Clock[0] = arg.Timestamp[0]
			}
			p.TimeStruct.Clock[0]++
			p.TimeStruct.Lock.Unlock()

			EnqueueMsg(arg, len(p.Membership), 2)

			res.Host = arg.Host
			res.Timestamp = make([]uint64, 1)
			res.Timestamp[0] = arg.Timestamp[0]
			res.Type = "ack"
			ChAck <- res
		} else {
			queueLock.Lock()
			submitAck(arg, len(p.Membership), 1)
			//for last < expected {
			//	var res utils.Sender
			//	datastoreHandler.coll.FindOne(datastoreHandler.ctx, bson.D{{"_id", last}}).Decode(&res)
			//	ChFromPeers <- res
			//	last++
			//}
			queueLock.Unlock()
		}
	}
	*res = 1

	return nil
}
