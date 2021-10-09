package utils

import "log"

type Sender struct {
	ID 			int		`bson:"_id"`
	Host 		string	`bson:"host"`
	Msg       	string	`bson:"msg"`
	Timestamp 	string	`bson:"order"`
	Type 		string
}

func ErrorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}