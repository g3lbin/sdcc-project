package rpcsequencer

import "fmt"

type Sequencer struct {
	Membership []string
}

type Sender struct {
	Host string
	Msg string
}

var sequenceNumber = 0

func (registry *Sequencer) SendInMulticast(arg Sender, res *int) error {
	fmt.Printf("##%d##\t%s has written: %s\n", sequenceNumber, arg.Host, arg.Msg)
	sequenceNumber++
	*res = 0

	return nil
}