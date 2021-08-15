package main

import (
	"log"
	"net"
	"net/rpc"
	"os"

	"app/lib"
)

func main() {
	rcv := new(lib.Receiver)

	// Register a new RPC server and the struct we created above
	server := rpc.NewServer()
	err := server.RegisterName("Receiver", rcv)
	if err != nil {
		log.Fatal("Format of service is not correct: ", err)
	}
	// Register an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	// server.HandleHTTP("/", "/debug")

	// Listen for incoming messages on port 4321
	lis, err := net.Listen("tcp", ":" + os.Args[1])
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	name, err := os.Hostname()
	if err != nil {
		log.Fatal("Oops: ", err)
	}

	rcv.Name = name

	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Fatal("Oops: ", err)
	}

	for _, a := range addrs {
		log.Printf("%s", a)
	}

	log.Printf("Registry on port %d", 4321)

	// Start go's http server on socket specified by listener
	// err = http.Serve(lis, nil)
	// if err != nil {
	// 	log.Fatal("Serve error: ", err)
	// }
	server.Accept(lis)
}
