package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"

	peer "github.com/sdcc-project/internal/pkg/rpcpeer"
	"github.com/sdcc-project/internal/pkg/utils"
)

var ip string

func retrieveIpAddr() string {
	var res string

	count := 0
	ifaces, err := net.Interfaces()
	if err != nil {
		utils.ErrorHandler("Interfaces", err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			utils.ErrorHandler("Addrs", err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.String() != "127.0.0.1" {
				res = ip.String()
				count++
			}
		}
	}
	if count != 1 {
		log.Fatal("There are more than one non-loopback interfaces")
	}

	return res
}

func registration() []string {
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

	msg = ip
	// Call remote procedure
	err = cl.Call("Registry.RegisterMember", msg, &list)
	if err != nil {
		utils.ErrorHandler("Call", err)
	}

	return list
}

func sendMessages(algo string, ch chan string) {
	var res int

	if algo == "tot-ordered-centr" {
		port, ok := os.LookupEnv("SEQUENCER_PORT")
		if !ok {
			log.Fatal("REGISTRATION_PORT environment variable is not set")
		}
		addr := "sequencer:" + port //address and port on which RPC server is listening

		sender := new(utils.Sender)
		sender.Host = ip
		sender.Msg = <-ch

		// Try to connect to addr
		cl, err := rpc.Dial("tcp", addr)
		if err != nil {
			utils.ErrorHandler("Dial", err)
		}
		for {
			// reply will store the RPC result
			// Call remote procedure
			err = cl.Call("Sequencer.SendInMulticast", sender, &res)
			if err != nil {
				utils.ErrorHandler("Call", err)
			}
			sender.Msg = <-ch
		}
	}
}

func getMessagesToSend(ch chan string) {
	var buffer string

	port, ok:= os.LookupEnv("CONTROL_PORT")
	if !ok {
		log.Fatal("CONTROL_PORT environment variable is not set")
	}

	// Listen for incoming messages on port CONTROL_PORT
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}
	conn, err := lis.Accept()
	if err != nil {
		utils.ErrorHandler("Accept", err)
	}
	reader := bufio.NewReader(conn)
	for {
		buffer, err = reader.ReadString('\n')
		ch <- buffer
	}
}

func getMessagesFromPeers() {
	var err error

	p := new(peer.Peer)

	// Register a new RPC server
	server := rpc.NewServer()
	err = server.RegisterName("Peer", p)
	if err != nil {
		utils.ErrorHandler("RegisterName", err)
	}

	lisPort, ok:= os.LookupEnv("MULTICAST_PORT")
	if !ok {
		log.Fatal("MULTICAST_PORT environment variable is not set")
	}

	// Listen for incoming messages on port LISTENING_PORT
	lis, err := net.Listen("tcp", ":" + lisPort)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}

	server.Accept(lis)
}

func main() {
	var algorithm string
	// var membership []string

	algorithm, ok := os.LookupEnv("MULTICAST_ALGORITHM")
	if !ok {
		log.Fatal("MULTICAST_ALGORITHM environment variable is not set")
	}

	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	membNum, err := strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}

	ip = retrieveIpAddr()

	//membership = registration()
	registration()

	go getMessagesFromPeers()
	peer.ChFromPeers = make(chan utils.Sender, membNum*10)
	chFromCL := make(chan string, 10)
	go getMessagesToSend(chFromCL)
	go sendMessages(algorithm, chFromCL)
	for {
		select {
		// case msgToSend := <- chFromCL:
		//	sendMessage(msgToSend, algorithm)
		 case receivedStruct := <- peer.ChFromPeers:
			 fmt.Printf("##%s##\t%s has written: %s", receivedStruct.Order, receivedStruct.Host, receivedStruct.Msg)
		}
	}
}
