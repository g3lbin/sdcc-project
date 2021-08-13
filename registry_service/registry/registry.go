package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"app/lib"
)

func main() {

	wr := new(lib.Writer)

	// Register a new RPC server and the struct we created above
	server := rpc.NewServer()
	err := server.RegisterName("Writer", wr)
	if err != nil {
		log.Fatal("Format of service is not correct: ", err)
	}
	// Register an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	// server.HandleHTTP("/", "/debug")

	// Listen for incoming messages on port 4321
	lis, err := net.Listen("tcp", ":4321")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	for _, a := range addrs {
		fmt.Println(a)
	}

	log.Printf("RPC server on port %d", 4321)

	// Start go's http server on socket specified by listener
	// err = http.Serve(lis, nil)
	// if err != nil {
	// 	log.Fatal("Serve error: ", err)
	// }
	server.Accept(lis)
}
