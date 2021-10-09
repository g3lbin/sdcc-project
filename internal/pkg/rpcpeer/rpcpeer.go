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
	Hostname string
	Algorithm string
	Membersihp []string
}

type Message struct {
	Timestamp struct {string; int}
	Type string
	RcvdAck int
	Content stringOrder
}

var queue []Message
var lock sync.RWMutex
var expected = 0
var ChFromPeers chan utils.Sender

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

	if p.Algorithm == "tot-ordered-centr" {
		chatCollection := client.Database(p.Hostname).Collection("chat")
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

	}
	*res = 1

	return nil
}
