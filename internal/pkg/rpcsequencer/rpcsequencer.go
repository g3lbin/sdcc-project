package rpcsequencer

import (
	"github.com/sdcc-project/internal/pkg/utils"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"sync"
)

type Sequencer struct {
	Membership []string
	Clients []*rpc.Client
	MembNum int
	SetupConn bool
}

var checkConnLock sync.Mutex
var seqNumLock sync.Mutex
var sequenceNumber = 0

func setupConnections(sequencer *Sequencer) {
	// Setup all connections
	port, ok := os.LookupEnv("PEER_PORT")
	if !ok {
		log.Fatal("PEER_PORT environment variable is not set")
	}
	for i := 0; i < sequencer.MembNum; i++ {
		addr := sequencer.Membership[i] + ":" + port // address and port on which RPC server is listening
		// Try to connect to addr
		cl, err := rpc.Dial("tcp", addr)
		if err != nil {
			utils.ErrorHandler("Dial", err)
		}
		sequencer.Clients = append(sequencer.Clients, cl)
	}
	sequencer.SetupConn = false
}

func (sequencer *Sequencer) SendInMulticast(arg utils.Sender, res *int) error {
	seqNumLock.Lock()
	arg.Order = strconv.Itoa(sequenceNumber)
	sequenceNumber++
	seqNumLock.Unlock()

	checkConnLock.Lock()
	if sequencer.SetupConn {
		setupConnections(sequencer)
	}
	checkConnLock.Unlock()
	for i := 0; i < sequencer.MembNum; i++ {
		// Call remote procedure
		err := sequencer.Clients[i].Call("Peer.ReceiveMessage", arg, res)
		if err != nil {
			utils.ErrorHandler("Call", err)
		}
	}

	*res = 0

	return nil
}