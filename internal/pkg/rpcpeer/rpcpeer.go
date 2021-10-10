package rpcpeer

import (
	"context"
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
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

type Message struct {
	ID uint64				`bson:"_id"`
	Timestamp struct {
		host string			`bson:"host"`
		clock []uint64		`bson:"clock"`
	}
	RcvdAck int
	Content string			`bson:"content"`
}

var queue []*Message
var queueLock sync.Mutex
var expected uint64 = 0
var last uint64 = 0
var ChFromPeers chan utils.Sender
var ChAck chan *utils.Sender

func EnqueueMsg(item utils.Sender) {
	newMsg := new(Message)
	newMsg.Timestamp.host = item.Host
	newMsg.Timestamp.clock = item.Timestamp
	newMsg.RcvdAck = 1
	newMsg.Content = item.Msg

	done := false
	for i, msg := range queue {
		if newMsg.Timestamp.clock[0] < msg.Timestamp.clock[0] ||
			(newMsg.Timestamp.clock[0] == msg.Timestamp.clock[0] && newMsg.Timestamp.host < msg.Timestamp.host) {
			tail := queue[i:]
			queue = append([]*Message{newMsg}, queue[:i]...)
			queue = append(queue, tail...)

			done = true
			break
		}
	}
	if !done {
		queue = append(queue, newMsg)
	}
}

func submitAck(ctx context.Context, coll *mongo.Collection, ack utils.Sender, membersNum int) {
	for i, msg := range queue {
		if msg.Timestamp.host == ack.Host && msg.Timestamp.clock[0] == ack.Timestamp[0] {
			msg.RcvdAck++
			if msg.RcvdAck == membersNum {
				msg.ID = expected
				expected++
				_, err := coll.InsertOne(ctx, msg)
				if err != nil {
					utils.ErrorHandler("InsertOne", err)
				}
				queue = append(queue[:i], queue[i+1:]...)
			}
		}
	}
}

func (p *Peer) ReceiveMessage(arg utils.Sender, res *int) error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	chatCollection := client.Database(p.Hostname).Collection("chat")

	if p.Algorithm == "tot-ordered-centr" {
		_, err = chatCollection.InsertOne(ctx, arg)
		if err != nil {
			utils.ErrorHandler("InsertOne", err)
		}
		for ok := true; ok; {
			var res utils.Sender
			err = chatCollection.FindOne(ctx, bson.D{{"_id", expected}}).Decode(&res)
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
		if arg.Type == "update" {
			// Update logical clock for 'receive' event
			p.TimeStruct.Lock.Lock()
			if p.TimeStruct.Clock[0] < arg.Timestamp[0] {
				p.TimeStruct.Clock[0] = arg.Timestamp[0]
			}
			p.TimeStruct.Clock[0]++
			p.TimeStruct.Lock.Unlock()

			queueLock.Lock()
			EnqueueMsg(arg)
			queueLock.Unlock()

			res := new(utils.Sender)
			res.Host = arg.Host
			res.Timestamp = arg.Timestamp
			res.Type = "ack"
			ChAck <- res
		} else {
			queueLock.Lock()
			submitAck(ctx, chatCollection, arg, len(p.Membership))
			for last < expected {
				var res utils.Sender
				chatCollection.FindOne(ctx, bson.D{{"_id", last}}).Decode(&res)
				ChFromPeers <- res
				last++
			}
			queueLock.Unlock()
		}
	}
	*res = 1

	return nil
}
