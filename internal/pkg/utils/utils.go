package utils

import "log"

type Sender struct {
	Host string
	Msg string
	Order string
}

func ErrorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}