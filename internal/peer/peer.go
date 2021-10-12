package main

import (
	"bufio"
	"fmt"
	peer "github.com/sdcc-project/internal/pkg/rpcpeer"
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

var ip string
var time *utils.Time

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

func sendMessages(algo string, chFromCL chan string, chAck chan utils.Sender, membership []string, membersNum int) {
	var sender utils.Sender
	var res int
	var err error

	select {
	case msg := <-chFromCL:
		sender.Host = ip
		sender.Msg = msg
		sender.Type = "update"
	case sender = <-chAck:
	}

	if algo == "tot-ordered-centr" {
		port, ok := os.LookupEnv("SEQUENCER_PORT")
		if !ok {
			log.Fatal("REGISTRATION_PORT environment variable is not set")
		}
		addr := "sequencer:" + port //address and port on which RPC server is listening

		// Try to connect to addr
		cl, err := rpc.Dial("tcp", addr)
		if err != nil {
			utils.ErrorHandler("Dial", err)
		}
		for {
			// reply will store the RPC result
			// Call remote procedure
			err = cl.Call("Sequencer.SendInMulticast", &sender, &res)
			if err != nil {
				utils.ErrorHandler("Call", err)
			}
			sender.Msg = <-chFromCL
		}
	} else if algo == "tot-ordered-decentr" {
		port, ok:= os.LookupEnv("MULTICAST_PORT")
		if !ok {
			log.Fatal("MULTICAST_PORT environment variable is not set")
		}

		clients := make([]*rpc.Client, membersNum - 1)
		i := 0
		for _, member := range membership {
			if member == ip {
				continue
			}
			clients[i], err = rpc.Dial("tcp", member + ":" + port)
			if err != nil {
				utils.ErrorHandler("Dial", err)
			}
			i++
		}

		for {
			if sender.Type == "update" {
				sender.Timestamp = make([]uint64, 1)
				// Update logical clock for 'send' event
				time.Lock.Lock()
				time.Clock[0]++
				sender.Timestamp[0] = time.Clock[0]
				time.Lock.Unlock()

				peer.EnqueueMsg(sender, membersNum, 1)			/* Send the message to the peer itself */
			}

			for i := 0; i < membersNum - 1; i++ {	/* Send the message to others */
				// Call remote procedure
				err = clients[i].Call("Peer.ReceiveMessage", &sender, &res)
				if err != nil {
					utils.ErrorHandler("Call", err)
				}
			}

			select {
			case msg := <-chFromCL:
				sender.Host = ip
				sender.Msg = msg
				sender.Type = "update"
			case sender = <-chAck:
			}
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

func getMessagesFromPeers(algo string, membership []string) {
	var err error

	p := new(peer.Peer)
	p.Hostname, err = os.Hostname()
	if err != nil {
		utils.ErrorHandler("Hostname", err)
	}
	p.Algorithm = algo
	p.Membership = membership
	p.TimeStruct = time

	peer.InitRpcPeer(ip)
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
	var membership []string

	time = new(utils.Time)
	algorithm, ok := os.LookupEnv("MULTICAST_ALGORITHM")
	if !ok {
		log.Fatal("MULTICAST_ALGORITHM environment variable is not set")
	}
	if algorithm == "tot-ordered-decentr" {
		time.Clock = append(time.Clock, 0)
	}

	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	membersNum, err := strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}

	ip = retrieveIpAddr()

	membership = registration()

	go getMessagesFromPeers(algorithm, membership)

	peer.ChFromPeers = make(chan utils.Sender, membersNum*10)
	peer.ChAck = make(chan utils.Sender, membersNum*10)
	chFromCL := make(chan string, 10)

	go getMessagesToSend(chFromCL)
	go sendMessages(algorithm, chFromCL, peer.ChAck, membership, membersNum)
	for {
		receivedStruct := <- peer.ChFromPeers
		fmt.Printf("#%-5s [%s] %s", receivedStruct.Timestamp, receivedStruct.Host, receivedStruct.Msg)
	}
}
