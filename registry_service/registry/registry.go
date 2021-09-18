package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"

	"app/lib"
)

func main() {
	var err error

	registry := new(lib.Registry)

	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	registry.MembersNum, err = strconv.Atoi(tmp)
	if err != nil {
		log.Fatal("Atoi error: ", err)
	}

	registry.FilePath = "/go/src/app/registry/registry.txt"

	// Register a new RPC server and the struct we created above
	server := rpc.NewServer()
	err = server.RegisterName("Registry", registry)
	if err != nil {
		log.Fatal("Format of service is not correct: ", err)
	}

	port, ok:= os.LookupEnv("LISTENING_PORT")
	if !ok {
		log.Fatal("LISTENING_PORT environment variable is not set")
	}

	// Listen for incoming messages on port LISTENING_PORT
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	log.Printf("Registry on port %s", port)
	server.Accept(lis)
}
