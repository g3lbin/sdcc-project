package rpcpeer

import (
	"context"
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Peer struct {
	Ip         string
	Algorithm  string
	Membership []string
	TimeStruct *utils.Time
	MembersId  map[string]int
}

var datastoreHandler struct { // maintain info to interact with the datastore
	ctx  context.Context
	coll *mongo.Collection
}
var queue []utils.Message   // local queue to maintain pending messages
var queueLock sync.Mutex    // lock to synchronize the accesses to queue variable
var deliveryWindow struct { // consecutive ID window of messages to be delivered to the application
	rightMargin uint64 // right margin of the delivery window
	leftMargin  uint64 // left margin of the delivery window
}
var ChUpdate chan utils.Message // channel to deliver update messages to the application

// InitRpcPeer initializes the global variables and establishes the connection with the datastore
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
	// create a database named as peer's hostname and create a collection named 'chat'
	// in the 'chat' collection will be stored all received messages
	datastoreHandler.coll = client.Database(hostname).Collection("chat")

	ackForMessages = make(map[string]int)
	messagesPerPeer = make(map[string]int)
	for _, member := range membership {
		messagesPerPeer[member] = 0
	}
	deliveryWindow.rightMargin = 0
	deliveryWindow.leftMargin = 0
}

// EnqueueMsg puts a pending message in the queue in according to the specified algorithm.
// The optional increment can be specified to indicate the number of acks to submit.
// The presence of the optional parameter depends on the algorithm.
func EnqueueMsg(algo string, new utils.Message, increment ...int) {
	queueLock.Lock()
	if algo == utils.ALGO2 {
		done := false
		inc := increment[1]
		// enqueue the new message in the right position
		for i, msg := range queue {
			if new.Timestamp[0] < msg.Timestamp[0] ||
				(new.Timestamp[0] == msg.Timestamp[0] && new.Host < msg.Host) {
				tail := append([]utils.Message{}, queue[i:]...)
				queue = append(queue[:i], new)
				queue = append(queue, tail...)

				done = true
				break
			}
		}
		// if the new message is the last one, append it at the end of queue
		if !done {
			queue = append(queue, new)
		}
		// increment the number of messages queued for new.Host
		if val, ok := messagesPerPeer[new.Host]; ok {
			messagesPerPeer[new.Host] = val + 1
		} else {
			messagesPerPeer[new.Host] = 1
		}
		// submit the first ack for the new message (the peer itself acknowledges the received message)
		submitAck(new, inc)
	} else if algo == utils.ALGO3 {
		queue = append(queue, new) // append the message to the queue
	}
	queueLock.Unlock()
}

// dequeue removes a pending message from the queue in according to the specified algorithm.
// The index parameter should be used to indicate the position in queue of the element to remove.
// The presence of the optional parameter depends on the algorithm.
func dequeue(algo string, index ...int) {
	if algo == utils.ALGO2 { // remove the message at the top of the queue
		head := queue[0]
		messagesPerPeer[head.Host]-- // decrement the number of messages queued for head.Host
		if len(queue) == 1 {
			queue = []utils.Message{}
		} else {
			queue = queue[1:]
		}
	} else if algo == utils.ALGO3 { // remove the specified message
		queue = append(queue[:index[0]], queue[index[0]+1:]...)
	}
}

// ReceiveMessage allows the caller to send a message to the peer p
func (p *Peer) ReceiveMessage(newMsg utils.Message, res *int) error {
	var err error

	switch p.Algorithm {
	case utils.ALGO1:
		err = receiveMessageAlgo1(newMsg)
	case utils.ALGO2:
		err = receiveMessageAlgo2(p.TimeStruct, p.Membership, newMsg)
	default:
		err = receiveMessageAlgo3(p.TimeStruct, len(p.Membership), p.MembersId, newMsg)
	}

	*res = 0

	return err
}
