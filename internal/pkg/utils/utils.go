package utils

import (
	"log"
	"sync"
)

type Sender struct {
	ID 			uint64	`bson:"_id"`
	Host 		string	`bson:"host"`
	Msg       	string	`bson:"msg"`
	Timestamp 	[]uint64
	Type 		string
}

type Time struct {
	Clock []uint64
	Lock  sync.Mutex
}

func ErrorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}