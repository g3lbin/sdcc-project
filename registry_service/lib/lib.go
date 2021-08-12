package lib

import "log"

type Writer int
type Message string
type Result int

func (wr *Writer) Write_on_terminal(arg Message, res *Result) error {

	log.Printf("Writer has received: %s\n", arg)
	*res = 777

	return nil
}
