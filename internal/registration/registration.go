package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"

	lib "github.com/sdcc-project/internal/pkg/rpcregistration"
	"github.com/sdcc-project/internal/pkg/utils"
)

var wg sync.WaitGroup

type regEnv struct {
	membersNum int
	listeningPort string
	algorithm string
}

// setEnvironment reads the environment variables and fill the regEnv struct with the read values
func setEnvironment(env *regEnv) {
	var err error
	var ok bool

	// retrieve number of multicast group members
	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	(*env).membersNum, err = strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}
	// retrieve port number of registration service
	(*env).listeningPort, ok = os.LookupEnv("LISTENING_PORT")
	if !ok {
		log.Fatal("LISTENING_PORT environment variable is not set")
	}
	// retrieve the type of the used multicast algorithm
	(*env).algorithm, ok = os.LookupEnv("MULTICAST_ALGORITHM")
	if !ok {
		log.Fatal("MULTICAST_ALGORITHM environment variable is not set")
	}
}

func main() {
	var requestsNum int
	var err error

	env := new(regEnv)
	setEnvironment(env)										// retrieve info from environment variables

	registry := new(lib.Registry)
	registry.MembersNum = env.membersNum
	registry.FilePath = "./registry.txt"

	// register a new RPC server
	server := rpc.NewServer()
	err = server.RegisterName("Registry", registry)
	if err != nil {
		utils.ErrorHandler("RegisterName", err)
	}

	// listen for incoming messages on port LISTENING_PORT
	lis, err := net.Listen("tcp", ":" + env.listeningPort)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}

	requestsNum = registry.MembersNum
	if env.algorithm == "tot-ordered-centr" {
		requestsNum++										// expect a request also from sequencer
	}

	fmt.Printf("Registration service on port %s...\n", env.listeningPort)
	wg.Add(requestsNum)										// create a waitgroup of requestNum members
	for i := 0; i < requestsNum; i++ {						// serve exactly requestNum requests
		go func() {
			conn, err := lis.Accept()
			if err != nil {
				utils.ErrorHandler("Accept", err)
			}
			server.ServeConn(conn)
			conn.Close()
			wg.Done()										// signal the end of the service
		}()
	}
	wg.Wait()												// wait for all connection and exit
}
