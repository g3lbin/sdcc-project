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
	MembNum int
}

var seqNumLock sync.Mutex
var sequenceNumber = 0

func (sequencer *Sequencer) SendInMulticast(arg utils.Sender, res *int) error {
	seqNumLock.Lock()
	sequenceNumber++
	arg.Order = strconv.Itoa(sequenceNumber)
	seqNumLock.Unlock()

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
		// Call remote procedure
		err = cl.Call("Peer.ReceiveMessage", arg, res)
		if err != nil {
			utils.ErrorHandler("Call", err)
		}
		cl.Close()
	}

	*res = 0

	return nil
}