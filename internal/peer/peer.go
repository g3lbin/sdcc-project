package main

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sdcc-project/internal/lib"
)

var name string
var wg sync.WaitGroup

func registration() []string {
	var msg lib.Message
	var list []string

	port, ok := os.LookupEnv("REGISTRY_PORT")
	if !ok {
		log.Fatal("REGISTRY_PORT environment variable is not set")
	}
	addr := "registry:" + port // address and port on which RPC server is listening
	// Try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer cl.Close()

	msg = lib.Message(name)
	// Call remote procedure
	err = cl.Call("Registry.RegisterMember", msg, &list)
	if err != nil {
		log.Fatal("Error in Registry.RegisterMember: ", err)
	}

	return list
}

func senderRoutine() {
	var msg lib.Message
	var res lib.Result

	defer wg.Done()

	rand.Seed(time.Now().UnixNano())
	min := 5
	max := 120

	addr := "registry:" + "4321" //address and port on which RPC server is listening
	// Try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer cl.Close()

	if len(os.Args) < 2 {
		log.Fatal("No args passed in")
	}

	msg = lib.Message(os.Args[1])

	for {
		// reply will store the RPC result
		// Call remote procedure
		err = cl.Call("Receiver.Receive_message", msg, &res)
		if err != nil {
			log.Fatal("Error in Receiver.Receive_message: ", err)
		}
		log.Printf("%s received: %d\n", name, res)
		res = 0

		r := rand.Intn(max - min + 1) + min
		msg = lib.Message(strconv.Itoa(r))
		time.Sleep(time.Duration(r) * time.Second)
	}
}

func receiverRoutine() {
	rcv := new(lib.Receiver)

	defer wg.Done()
	// Register a new RPC server and the struct we created above
	server := rpc.NewServer()
	err := server.RegisterName("Receiver", rcv)
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

	rcv.Name = name

	//addrs, err := net.LookupHost(name)
	//if err != nil {
	//	fmt.Printf("Oops: %v\n", err)
	//	return
	//}
	//
	//for _, a := range addrs {
	//	fmt.Println(a)
	//}

	// Start go's http server on socket specified by listener
	// err = http.Serve(lis, nil)
	// if err != nil {
	// 	log.Fatal("Serve error: ", err)
	// }
	server.Accept(lis)
}

func main() {
	var membership []string
	var err error

	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	membNum, err := strconv.Atoi(tmp)
	if err != nil {
		log.Fatal("Atoi error: ", err)
	}

	name, err = os.Hostname()
	if err != nil {
		log.Fatal("Hostname error: ", err)
	}

	membership = registration()
	for i := 0; i < membNum; i++ {
		log.Printf("%s\n", membership[i])
	}

	//wg = sync.WaitGroup{}
	//wg.Add(2)
	//go senderRoutine()
	//go receiverRoutine()
	//wg.Wait()
}
