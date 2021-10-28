package rpcsequencer

import (
	"github.com/sdcc-project/internal/pkg/utils"
	"net/rpc"
	"sync"
)

type SeqEnv struct {
	MembersNum int
	RegistrationPort string
	ListeningPort string
	PeerPort string
}

type Sequencer struct {
	Membership []string
	Clients    []*rpc.Client
	Env SeqEnv
}

var doOnce sync.Once							// execute a statement only first time (concurrency safe)
var seqNumLock sync.Mutex						// lock to synchronize accesses to sequenceNumber
var sequenceNumber uint64 = 0					// sequence number to tag messages

// setupConnections establishes the connections with all peers. This function should be run only once
func setupConnections(sequencer *Sequencer) {
	// Setup all connections
	for i := 0; i < sequencer.Env.MembersNum; i++ {
		addr := sequencer.Membership[i] + ":" + sequencer.Env.PeerPort // address and port on which RPC server is listening
		// Try to connect to addr
		cl, err := rpc.Dial("tcp", addr)
		if err != nil {
			utils.ErrorHandler("Dial", err)
		}
		sequencer.Clients = append(sequencer.Clients, cl)
	}
}

// SendInMulticast sends received message to all peers
func (sequencer *Sequencer) SendInMulticast(arg utils.Sender, res *int) error {
	seqNumLock.Lock()
	arg.ID = sequenceNumber
	arg.Timestamp = make([]uint64, 1)
	arg.Timestamp[0] = sequenceNumber
	sequenceNumber++
	seqNumLock.Unlock()

	doOnce.Do(func() {						// setup connections only first time
		setupConnections(sequencer)
	})
	// send received message to all peers
	for i := 0; i < sequencer.Env.MembersNum; i++ {
		// call remote procedure
		err := sequencer.Clients[i].Call("Peer.ReceiveMessage", arg, res)
		if err != nil {
			utils.ErrorHandler("Call", err)
		}
	}

	return nil
}