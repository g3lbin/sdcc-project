package utils

import "log"

func ErrorHandler(foo string, err error) {
	log.Fatalf("%s has failed: %s", foo, err)
}