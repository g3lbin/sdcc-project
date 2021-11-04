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
	"strings"
)

type peerEnv struct {
	membersNum       int
	delay            int
	algorithm        string
	registrationPort string
	multicastPort    string
	controlPort      string
	sequencerPort    string
}

var membersId map[string]int // map to identify the peers using a numerical id
var time *utils.Time         // struct to maintain current logical time

// setEnvironment reads the environment variables and fill the peerEnv struct with the read values
func setEnvironment(env *peerEnv) {
	var err error
	var ok bool

	// retrieve the number of the multicast group members
	tmp, ok := os.LookupEnv("MEMBERS_NUM")
	if !ok {
		log.Fatal("MEMBERS_NUM environment variable is not set")
	}
	(*env).membersNum, err = strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}
	// retrieve the type of the used multicast algorithm
	(*env).algorithm, ok = os.LookupEnv("MULTICAST_ALGORITHM")
	if !ok {
		log.Fatal("MULTICAST_ALGORITHM environment variable is not set")
	}
	// retrieve the port number of registration service
	(*env).registrationPort, ok = os.LookupEnv("REGISTRATION_PORT")
	if !ok {
		log.Fatal("REGISTRATION_PORT environment variable is not set")
	}
	// retrieve the port number of peer service to send and receive multicast messages
	(*env).multicastPort, ok = os.LookupEnv("MULTICAST_PORT")
	if !ok {
		log.Fatal("MULTICAST_PORT environment variable is not set")
	}
	// retrieve the port number to receive messages from CLI
	(*env).controlPort, ok = os.LookupEnv("CONTROL_PORT")
	if !ok {
		log.Fatal("CONTROL_PORT environment variable is not set")
	}
	// retrieve the port number of sequencer service
	(*env).sequencerPort, ok = os.LookupEnv("SEQUENCER_PORT")
	if !ok {
		log.Fatal("SEQUENCER_PORT environment variable is not set")
	}
	// retrieve the delay parameter
	tmp, ok = os.LookupEnv("DELAY")
	if !ok {
		log.Fatal("DELAY environment variable is not set")
	}
	(*env).delay, err = strconv.Atoi(tmp)
	if err != nil {
		utils.ErrorHandler("Atoi", err)
	}
}

// retrieveIpAddr loops through all network interfaces to find the peer's IP address
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
			if ip.String() != "127.0.0.1" { // found an IP address of a non-loopback interface
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

// registration connects to the registration service to register the peer's address and to get the list of members
func registration(targetPort string, myIpAddr string) []string {
	var list []string

	addr := "registration:" + targetPort // address and port on which RPC server is listening
	// try to connect to addr
	cl, err := rpc.Dial("tcp", addr)
	if err != nil {
		utils.ErrorHandler("Dial", err)
	}
	defer cl.Close()

	// call remote procedure
	err = cl.Call("Registry.RegisterMember", myIpAddr, &list)
	if err != nil {
		utils.ErrorHandler("Call", err)
	}

	return list
}

// sendMessages takes "update messages" from the CLI and "ack messages" from peer.ReceiveMessage using the channels
// and multicasts them, according to the specified algorithm
func sendMessages(
	algo string, // multicast algorithm
	serverPort string, // the port of the target to send the message to
	chCLI chan string, // channel to receive, from the CLI, the update messages to send to other peers
	chAck chan utils.Message, // channel to receive the ack messages to send to other peers
	membership []string, // list of all peer addresses
	membersNum int, // number of peers
	myIpAddr string, // IP address of the peer (the sender in this case)
	delay int, // delay parameter to postpone the sending of messages (used to test)
) {
	var message utils.Message
	var res int
	var chPerPeer []chan utils.Message // channels to send messages to each rpcHandler

	// wait for the first message to send
	select {
	case msg := <-chCLI:
		message.Host = myIpAddr
		message.Content = msg
		message.Type = utils.UPDATE
	case message = <-chAck:
	}

	if algo == utils.ALGO1 {
		addr := "sequencer:" + serverPort //address and port on which RPC server is listening

		// try to connect to addr
		cl, err := rpc.Dial("tcp", addr)
		if err != nil {
			utils.ErrorHandler("Dial", err)
		}
		// infinite loop to send messages to sequencer service
		for {
			// call remote procedure
			err = cl.Call("Sequencer.SendInMulticast", &message, &res)
			if err != nil {
				utils.ErrorHandler("Call", err)
			}
			message.Content = <-chCLI // get the message from the CLI
		}
	} else {
		// establish the connections with other peers
		chPerPeer = make([]chan utils.Message, membersNum-1)
		i := 0
		for _, member := range membership {
			if member == myIpAddr {
				continue
			}
			chPerPeer[i] = make(chan utils.Message, 10*membersNum)
			serviceAddr := member + ":" + serverPort // address and port on which RPC server is listening
			go utils.RpcHandler(serviceAddr, "Peer.ReceiveMessage", chPerPeer[i], delay)
			i++
		}
		// infinite loop to send multicast messages to other peers
		for {
			if message.Type == utils.UPDATE && algo == utils.ALGO2 {
				message.Timestamp = make([]uint64, 1)
				// update logical scalar clock for 'send' event
				time.Lock.Lock()
				time.Clock[0]++
				message.Timestamp[0] = time.Clock[0] // tag the message using the clock as timestamp
				// update logical clock for 'receive' event
				time.Clock[0]++
				time.Lock.Unlock()
				peer.EnqueueMsg(algo, message, membersNum, 1) // put the message in the local queue
			} else if algo == utils.ALGO3 {
				message.Timestamp = make([]uint64, membersNum)
				// update vector logical clock for 'send' event
				time.Lock.Lock()
				time.Clock[membersId[myIpAddr]]++ // increment the vector's entry which correspond to the peer
				for i := 0; i < membersNum; i++ { // tag the message using the clock as timestamp
					message.Timestamp[i] = time.Clock[i]
				}
				peer.VectDelivery(message, time, membersId) // deliver the message to the peer itself
				time.Lock.Unlock()
			}
			// send the message to others
			for i := 0; i < membersNum-1; i++ {
				chPerPeer[i] <- message
			}

			// wait for a message to send
			select {
			case msg := <-chCLI:
				message.Host = myIpAddr
				message.Content = msg
				message.Type = utils.UPDATE
			case message = <-chAck:
			}
		}
	}
}

// getMessagesToSend takes messages from the CLI and sends them to sendMessages using ch
func getMessagesToSend(ch chan string, port string) {
	var buffer string

	// listen for incoming messages on port CONTROL_PORT
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}
	// establish a connection with the CLI
	conn, err := lis.Accept()
	if err != nil {
		utils.ErrorHandler("Accept", err)
	}
	// loop to read the messages received by CLI
	reader := bufio.NewReader(conn)
	for {
		buffer, err = reader.ReadString('\n')
		if err != nil {
			utils.ErrorHandler("ReadString", err)
		}
		ch <- buffer
	}
}

// getMessagesFromPeers registers the RPC server to receive multicast messages
func getMessagesFromPeers(algo string, myPort string, myIpAddr string, membership []string) {
	var err error

	p := new(peer.Peer)
	hostname, err := os.Hostname()
	if err != nil {
		utils.ErrorHandler("Hostname", err)
	}
	p.Ip = myIpAddr
	p.Algorithm = algo
	p.Membership = membership
	p.TimeStruct = time
	p.MembersId = membersId

	peer.InitRpcPeer(hostname, membership) // set up the necessary structures
	// register a new RPC server
	server := rpc.NewServer()
	err = server.RegisterName("Peer", p)
	if err != nil {
		utils.ErrorHandler("RegisterName", err)
	}

	// listen for incoming messages on port MULTICAST_PORT
	lis, err := net.Listen("tcp", ":"+myPort)
	if err != nil {
		utils.ErrorHandler("Listen", err)
	}

	server.Accept(lis)
}

func main() {
	var env peerEnv
	var serverPort string // the port of the target to send the message to
	var myIpAddr string   // the IP address of the peer
	var membership []string

	setEnvironment(&env) // retrieve info from environment variables
	time = new(utils.Time)

	if env.algorithm == utils.ALGO2 {
		time.Clock = append(time.Clock, 0)                       // scalar logical clock
		peer.ChAck = make(chan utils.Message, env.membersNum*10) // create a channel for ack messages
	} else if env.algorithm == utils.ALGO3 {
		time.Clock = make([]uint64, env.membersNum) // vector logical clock
	}

	myIpAddr = retrieveIpAddr() // get the IP address

	membership = registration(env.registrationPort, myIpAddr) // register the peer and get the membership
	// assign a numerical id to the peers
	membersId = make(map[string]int)
	for i, member := range membership {
		membersId[member] = i
	}

	peer.ChUpdate = make(chan utils.Message, env.membersNum*10) // create a channel for update messages
	chCLI := make(chan string, 10)                              // create a channel for CLI messages

	go getMessagesFromPeers(env.algorithm, env.multicastPort, myIpAddr, membership) // start to receive message from others
	go getMessagesToSend(chCLI, env.controlPort)                                    // start to receive message from CLI

	// select the port to send messages to
	if env.algorithm == utils.ALGO1 {
		serverPort = env.sequencerPort
	} else {
		serverPort = env.multicastPort
	}
	// get ready to send messages
	go sendMessages(env.algorithm, serverPort, chCLI, peer.ChAck, membership, env.membersNum, myIpAddr, env.delay)

	// print loop for received messages
	padding := strconv.Itoa(2*env.membersNum+env.membersNum-1)
	for {
		deliveredMsg := <-peer.ChUpdate
		order := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(deliveredMsg.Timestamp)), ","), "[]")
		fmt.Printf("<%-"+padding+"s [%s] %s", order+">", deliveredMsg.Host, deliveredMsg.Content)
	}
}
