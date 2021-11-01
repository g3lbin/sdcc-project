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

// setEnvironment reads the environment variables and fill the seqEnv struct with the read values
func setEnvironment(env *lib.SeqEnv) {
	var err error
	var ok bool

	// retrieve the number of the multicast group members
	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	(*env).MembersNum, err = strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}
	// retrieve the delay parameter
	tmp, ok = os.LookupEnv("DELAY")
	if !ok {
		log.Fatal("DELAY environment variable is not set")
	}
	(*env).Delay, err = strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}
	// retrieve port number of registration service
	(*env).RegistrationPort, ok = os.LookupEnv("REGISTRATION_PORT")
	if !ok {
		log.Fatal("REGISTRATION_PORT environment variable is not set")
	}
	// retrieve port number of sequencer service
	(*env).ListeningPort, ok = os.LookupEnv("LISTENING_PORT")
	if !ok {
		log.Fatal("LISTENING_PORT environment variable is not set")
	}
	// retrieve port number of peer service
	(*env).PeerPort, ok = os.LookupEnv("PEER_PORT")
	if !ok {
		log.Fatal("PEER_PORT environment variable is not set")
	}
}

// getMembership returns membership retrieved from registration service
func getMembership(port string) []string {
	var msg string
	var list []string

	addr := "registration:" + port // address and port on which RPC server is listening
	// Try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		utils.ErrorHandler("Dial", err)
	}
	defer cl.Close()

	msg = "" // no message to send to registration service
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
	setEnvironment(&sequencer.Env)                                       // retrieve info from environment variables
	sequencer.Membership = getMembership(sequencer.Env.RegistrationPort) // get the membership from registration service

	// Register a new RPC server
	server := rpc.NewServer()
	err = server.RegisterName("Sequencer", sequencer)
	if err != nil {
		utils.ErrorHandler("RegisterName", err)
	}

	// Listen for incoming messages on port LISTENING_PORT
	lis, err := net.Listen("tcp", ":"+sequencer.Env.ListeningPort)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}

	fmt.Printf("Sequencer service on port %s...\n", sequencer.Env.ListeningPort)
	server.Accept(lis)
}
