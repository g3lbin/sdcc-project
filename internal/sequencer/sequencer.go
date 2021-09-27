package main

import (
	"fmt"
	lib "github.com/sdcc-project/internal/pkg/rpcsequencer"
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

func getMembership() []string {
	var msg string
	var list []string

	port, ok := os.LookupEnv("REGISTRATION_PORT")
	if !ok {
		log.Fatal("REGISTRATION_PORT environment variable is not set")
	}
	addr := "registration:" + port // address and port on which RPC server is listening
	// Try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		utils.ErrorHandler("Dial", err)
	}
	defer cl.Close()

	msg = "get me the list of members"
	// Call remote procedure
	err = cl.Call("Registry.RetrieveMembership", msg, &list)
	if err != nil {
		utils.ErrorHandler("Call", err)
	}

	return list
}

func main() {
	var err error

	sequencer := new(lib.Sequencer)
	sequencer.Membership = getMembership()

	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	sequencer.MembNum, err = strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}

	// Register a new RPC server
	server := rpc.NewServer()
	err = server.RegisterName("Sequencer", sequencer)
	if err != nil {
		utils.ErrorHandler("RegisterName", err)
	}

	lisPort, ok:= os.LookupEnv("LISTENING_PORT")
	if !ok {
		log.Fatal("LISTENING_PORT environment variable is not set")
	}

	// Listen for incoming messages on port LISTENING_PORT
	lis, err := net.Listen("tcp", ":" + lisPort)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}

	fmt.Printf("Sequencer service on port %s...\n", lisPort)
	server.Accept(lis)
}
