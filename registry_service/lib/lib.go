package lib

import "log"

type Receiver struct {
	Name string
	Str Message
}
type Message string
type Result int

func (rcv *Receiver) Receive_message(arg Message, res *Result) error {

	rcv.Str = arg
	log.Printf("%s received: %s", rcv.Name, rcv.Str)
	*res = 777

	return nil
}
