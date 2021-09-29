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

func main() {
	var requestsNum int
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

	algorithm, ok := os.LookupEnv("MULTICAST_ALGORITHM")
	if !ok {
		log.Fatal("MULTICAST_ALGORITHM environment variable is not set")
	}
	requestsNum = registry.MembersNum
	if algorithm == "tot-ordered-centr" {
		requestsNum++
	}

	fmt.Printf("Registration service on port %s...\n", port)
	wg.Add(requestsNum)
	for i := 0; i < requestsNum; i++ {
		go func() {
			conn, err := lis.Accept()
			if err != nil {
				utils.ErrorHandler("Accept", err)
			}
			server.ServeConn(conn)
			conn.Close()
			wg.Done()
		}()
	}
	wg.Wait()
}
