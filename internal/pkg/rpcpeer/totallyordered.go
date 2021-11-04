package rpcpeer

import (
	"fmt"
	"github.com/sdcc-project/internal/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

var ChAck chan utils.Message // channel to send ack messages to routine sendMessages
// following variables are used to optimize the handling of the sorted queue
var ackForMessages map[string]int  // map to count the number of received ack for each message
var messagesPerPeer map[string]int // map to count the number of messages queued for each peer

// checkMsgForEachPeer returns true only if there is a queued message sent by each peer, false otherwise.
func checkMsgForEachPeer(membership []string) bool {
	for _, member := range membership {
		if messagesPerPeer[member] == 0 {
			return false
		}
	}

	return true
}

// submitAck adds increment to the number of received ack for a pending message.
// The increment parameter should be 2 to acknowledge a new message sent by another peer:
// 1 ack from the sender and 1 ack from the receiver.
func submitAck(ack utils.Message, increment int) {
	// the message is identified by a string structured as "hostname:timestamp"
	msgId := ack.Host + ":" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ack.Timestamp)), ","), "[]")
	if val, ok := ackForMessages[msgId]; ok {
		ackForMessages[msgId] = val + increment
	} else {
		ackForMessages[msgId] = increment
	}
}

// checkForDelivery scans the message queue and checks if there are any messages that can be delivered to the application.
// If so, proceed with the delivery.
func checkForDelivery(membership []string) {
	var headOfQueue utils.Message

	// loop to remove those messages from the queue and put them in the datastore
	for {
		queueLock.Lock()
		if len(queue) != 0 {
			headOfQueue = queue[0]
		} else {
			break
		}
		// check if all acks for the message at the top of the queue have been received and
		// if there is at least one message for each peer in the queue
		headId := headOfQueue.Host + ":" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(headOfQueue.Timestamp)), ","), "[]")
		if ackForMessages[headId] == len(membership) && checkMsgForEachPeer(membership) {
			dequeue(utils.ALGO2)           // pop the head of the queue
			delete(ackForMessages, headId) // remove the pair (msg,#acks) from the map

			headOfQueue.ID = deliveryWindow.rightMargin // set the ID of the deliverable message
			deliveryWindow.rightMargin++                // increment the expected message index
			// insert headOfQueue msg in the collection
			_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, headOfQueue)
			if err != nil {
				utils.ErrorHandler("InsertOne", err)
			}
		} else {
			break
		}
		queueLock.Unlock()
	}
	// loop to deliver all messages whose indexes are present in the deliveryWindow
	for deliveryWindow.leftMargin < deliveryWindow.rightMargin {
		var delivery utils.Message
		datastoreHandler.coll.FindOne(datastoreHandler.ctx, bson.D{{"_id", deliveryWindow.leftMargin}}).Decode(&delivery)
		ChUpdate <- delivery        // deliver the message to the application
		deliveryWindow.leftMargin++ // move the left margin of the window to the right
	}
	queueLock.Unlock()
}

// receiveMessageAlgo2 is the implementation of the ReceiveMessage method body according to algorithm 2
func receiveMessageAlgo2(time *utils.Time, membership []string, newMsg utils.Message) error {
	var ack utils.Message

	if newMsg.Type == utils.UPDATE {
		// update logical clock for 'receive' event
		time.Lock.Lock()
		if time.Clock[0] < newMsg.Timestamp[0] {
			time.Clock[0] = newMsg.Timestamp[0]
		}
		time.Clock[0]++
		time.Lock.Unlock()

		// queue the received message and submit 2 acks for it: 1 from the sender and 1 from the receiver
		EnqueueMsg(utils.ALGO2, newMsg, len(membership), 2)

		// build an ack message to send to all others peer
		ack.Host = newMsg.Host
		ack.Timestamp = make([]uint64, 1)
		ack.Timestamp[0] = newMsg.Timestamp[0]
		ack.Type = utils.ACK
		ChAck <- ack // send the ack to sendMessages routine
	} else {
		// an ack message was received, so increment the ack number for the relevant message
		queueLock.Lock()
		submitAck(newMsg, 1)
		queueLock.Unlock()
	}
	// check if there are any messages to deliver to the application
	checkForDelivery(membership)

	return nil
}
