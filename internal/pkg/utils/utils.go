package utils

import (
	"log"
	"math/rand"
	"net/rpc"
	"sync"
	t "time"
)

const (
	ALGO1  = "centralized_totally_ordered_multicast"
	ALGO2  = "decentralized_totally_ordered_multicast"
	ALGO3  = "causally_ordered_multicast"
	UPDATE = "update_msg"
	ACK    = "ack_msg"
)

type Message struct {
	ID        uint64   `bson:"_id"`  // message ID to retrieve ordinated documents from the datastore
	Host      string   `bson:"host"` // host related to message
	Content   string   `bson:"msg"`  // message content
	Timestamp []uint64 // value that tag the message
	Type      string   // type of message (update or ack)
}

// Time allows to maintain and manage the current logical time
type Time struct {
	Clock []uint64
	Lock  sync.Mutex
}

// ErrorHandler allows to handle the errors according to the preferred policy
func ErrorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}

// RpcHandler establishes a connection with an RPC server and calls in an infinity loop a specified RPC method.
// This method should be called inside a goroutine (e.g. go RpcHandler...)
func RpcHandler(serviceAddress string, serviceMethod string, ch chan Message, delay int) {
	var msg Message
	var res int

	client, err := rpc.Dial("tcp", serviceAddress)
	if err != nil {
		ErrorHandler("Dial", err)
	}
	// infinite loop to send messages to the remote peer
	for {
		msg = <-ch
		if delay > 0 {
			r := rand.Intn(delay + 1)
			t.Sleep(t.Duration(r) * t.Second)
		}
		// call remote procedure
		err = client.Call(serviceMethod, &msg, &res)
		if err != nil {
			ErrorHandler("Call", err)
		}
	}
}
