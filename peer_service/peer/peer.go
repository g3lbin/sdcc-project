package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	"app/lib"
)

func main() {
	var msg lib.Message
	var res lib.Result

	addr := "10.5.0.2:" + "4321" //address and port on which RPC server is listening
	// Try to connect to addr
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer client.Close()

	if len(os.Args) < 2 {
		fmt.Printf("No args passed in\n")
		os.Exit(1)
	}

	msg = lib.Message(os.Args[1])

	// reply will store the RPC result
	// Call remote procedure
	log.Printf("Synchronous call to RPC server")
	err = client.Call("Writer.Write_on_terminal", msg, &res)
	if err != nil {
		log.Fatal("Error in Writer.Write_on_terminal: ", err)
	}

	fmt.Printf("Writer.Write_on_terminal result: %d\n\n", res)

	// Asynchronous call
	// divReply := new(arith.Quotient)
	// log.Printf("Asynchronous call to RPC server")
	// divCall := client.Go("Arithmetic.Divide", args, divReply, nil)
	// divCall = <-divCall.Done
	// if divCall.Error != nil {
	// 	log.Fatal("Error in Arithmetic.Divide: ", divCall.Error.Error())
	// }
	// fmt.Printf("Arithmetic.Divide: %d/%d=%d (rem=%d)\n", args.A, args.B, divReply.Quo, divReply.Rem)
}
