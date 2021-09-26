package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"

	lib "github.com/sdcc-project/internal/pkg/rpcregistration"
	"github.com/sdcc-project/internal/pkg/utils"
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
		utils.ErrorHandler("Atoi", err)
	}

	registry.FilePath = "./registry.txt"

	// Register a new RPC server and the struct we created above
	server := rpc.NewServer()
	err = server.RegisterName("Registry", registry)
	if err != nil {
		utils.ErrorHandler("RegisterName", err)
	}

	port, ok:= os.LookupEnv("LISTENING_PORT")
	if !ok {
		log.Fatal("LISTENING_PORT environment variable is not set")
	}

	// Listen for incoming messages on port LISTENING_PORT
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}

	fmt.Printf("Registration service on port %s...\n", port)
	server.Accept(lis)
}
