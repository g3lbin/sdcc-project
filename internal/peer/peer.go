package main

import (
	"bufio"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sdcc-project/internal/pkg/lib"
)

var name string
var wg sync.WaitGroup

func registration() []string {
	var msg lib.Message
	var list []string

	port, ok := os.LookupEnv("REGISTRATION_PORT")
	if !ok {
		log.Fatal("REGISTRATION_PORT environment variable is not set")
	}
	addr := "registration:" + port // address and port on which RPC server is listening
	// Try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		lib.ErrorHandler("Dial", err)
	}
	defer cl.Close()

	msg = lib.Message(name)
	// Call remote procedure
	err = cl.Call("Registry.RegisterMember", msg, &list)
	if err != nil {
		lib.ErrorHandler("Call", err)
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

	addr := "registration:" + "4321" //address and port on which RPC server is listening
	// Try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		lib.ErrorHandler("Dial", err)
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
			lib.ErrorHandler("Call", err)
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
		lib.ErrorHandler("RegisterName", err)
	}
	// Register an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	// server.HandleHTTP("/", "/debug")

	// Listen for incoming messages on port 4321
	lis, err := net.Listen("tcp", ":4321")
	if err != nil {
		lib.ErrorHandler("Listen", err)
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

func getMessage() {
	var buffer string

	defer wg.Done()

	port, ok:= os.LookupEnv("CONTROL_PORT")
	if !ok {
		log.Fatal("CONTROL_PORT environment variable is not set")
	}

	// Listen for incoming messages on port CONTROL_PORT
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		lib.ErrorHandler("Listen", err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		clientReader := bufio.NewReader(conn)
		buffer, err = clientReader.ReadString('\n')
		log.Printf("%s has received: %s", name, buffer)
		conn.Close()
	}
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
		lib.ErrorHandler("Atoi", err)
	}

	name, err = os.Hostname()
	if err != nil {
		lib.ErrorHandler("Hostname", err)
	}

	membership = registration()
	for i := 0; i < membNum; i++ {
		log.Printf("%s\n", membership[i])
	}

	wg = sync.WaitGroup{}
	wg.Add(1)
	//go senderRoutine()
	//go receiverRoutine()
	go getMessage()
	wg.Wait()
}
