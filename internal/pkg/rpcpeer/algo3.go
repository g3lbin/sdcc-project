package rpcpeer

import "github.com/sdcc-project/internal/pkg/utils"

func VectDelivery(new utils.Message, time *utils.Time, membersId map[string]int) {
	queueLock.Lock()
	new.ID = deliveryWindow.rightMargin // set the ID of the deliverable message
	deliveryWindow.rightMargin++        // increment the expected message index
	queueLock.Unlock()

	// insert the new msg in the collection
	_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, new)
	if err != nil {
		utils.ErrorHandler("InsertOne", err)
	}
	ChUpdate <- new // deliver the message to the application

	// check for other messages to deliver
	for ok := true; ok; ok = chainDeliveries(time, membersId) {
	}
}

// chainDeliveries checks whether a pending message has become deliverable due to the previous delivery.
// The function returns "true" if at least one message has been delivered, "false" otherwise.
// The caller should call this function in a loop until the result becomes "false".
func chainDeliveries(time *utils.Time, membersId map[string]int) bool {
	var res = false

	queueLock.Lock()
	for j, msg := range queue {
		acceptableMsg := true
		time.Lock.Lock()
		// check if there are any events that the sender of msg has seen but the receiver has not yet
		for i := 0; i < len(membersId); i++ {
			if (i != membersId[msg.Host] && msg.Timestamp[i] > time.Clock[i]) ||
				(i == membersId[msg.Host] && msg.Timestamp[i] != time.Clock[i]+1) {
				acceptableMsg = false
				break
			}
		}
		if acceptableMsg {
			time.Clock[membersId[msg.Host]]++ // increment the vector's entry which correspond to the sender
		}
		time.Lock.Unlock()

		msg.ID = deliveryWindow.rightMargin // set the ID of the deliverable message
		deliveryWindow.rightMargin++        // increment the expected message index
		// insert the new msg in the collection
		_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, msg)
		if err != nil {
			utils.ErrorHandler("InsertOne", err)
		}
		dequeue(utils.ALGO3, j) // pop the deliverable message from the queue
		ChUpdate <- msg         // deliver the message to the application

		res = true
	}
	queueLock.Unlock()

	return res
}

// receiveMessageAlgo3 is the implementation of the ReceiveMessage method body according to algorithm 3
func receiveMessageAlgo3(time *utils.Time, membersNum int, membersId map[string]int, newMsg utils.Message) error {
	// check if the received message can be delivered to the application
	acceptableMsg := true
	time.Lock.Lock()
	for i := 0; i < membersNum; i++ {
		// check if there are any events that the sender has seen but the receiver has not yet
		if (i != membersId[newMsg.Host] && newMsg.Timestamp[i] > time.Clock[i]) ||
			(i == membersId[newMsg.Host] && newMsg.Timestamp[i] != time.Clock[i]+1) {
			acceptableMsg = false
			break
		}
	}
	if acceptableMsg {
		time.Clock[membersId[newMsg.Host]]++ // increment the vector's entry which correspond to the sender
		time.Lock.Unlock()
		VectDelivery(newMsg, time, membersId) // deliver the message to the application
	} else {
		time.Lock.Unlock()
		EnqueueMsg(utils.ALGO3, newMsg) // append the message to the queue
	}

	return nil
}
