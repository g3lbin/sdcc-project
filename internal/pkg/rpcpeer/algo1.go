package rpcpeer

import (
	"github.com/sdcc-project/internal/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// receiveMessageAlgo1 is the implementation of the ReceiveMessage method body according to algorithm 1
func receiveMessageAlgo1(newMsg utils.Message) error {
	// insert the new msg in the collection
	_, err := datastoreHandler.coll.InsertOne(datastoreHandler.ctx, newMsg)
	if err != nil {
		utils.ErrorHandler("InsertOne", err)
	}
	// check if there are any messages to deliver to the application
	for ok := true; ok; {
		var delivery utils.Message
		// retrieve the first undelivered item according to deliveryWindow.rightMargin from the collection.
		// deliveryWindow.leftMargin should not be used for this algorithm
		err = datastoreHandler.coll.FindOne(
			datastoreHandler.ctx,
			bson.D{{"_id", deliveryWindow.rightMargin}},
		).Decode(&delivery)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				ok = false
			} else {
				utils.ErrorHandler("FindOne", err)
			}
		} else {
			ChUpdate <- delivery         // deliver the message to the application
			deliveryWindow.rightMargin++ // increment the expected message index
		}
	}

	return nil
}
