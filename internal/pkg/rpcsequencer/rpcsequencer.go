package rpcsequencer

import (
	"github.com/sdcc-project/internal/pkg/utils"
	"sync"
)

type SeqEnv struct {
	MembersNum       int
	Delay            int
	RegistrationPort string
	ListeningPort    string
	PeerPort         string
}

type Sequencer struct {
	Membership []string
	Env        SeqEnv
	chPerPeer []chan utils.Message // channels to send messages to each rpcHandler
}

var doOnce sync.Once          // execute a statement only first time (concurrency safe)
var seqNumLock sync.Mutex     // lock to synchronize accesses to sequenceNumber
var sequenceNumber uint64 = 0 // sequence number to tag messages

// SendInMulticast sends received message to all peers
func (sequencer *Sequencer) SendInMulticast(arg utils.Message, res *int) error {
	seqNumLock.Lock()
	arg.ID = sequenceNumber
	arg.Timestamp = make([]uint64, 1)
	arg.Timestamp[0] = sequenceNumber
	sequenceNumber++
	seqNumLock.Unlock()

	// establish the connection with all peers (only the first time)
	doOnce.Do(func() {
		sequencer.chPerPeer = make([]chan utils.Message, sequencer.Env.MembersNum)
		for i := 0; i < sequencer.Env.MembersNum; i++ {
			serviceAddr := sequencer.Membership[i] + ":" + sequencer.Env.PeerPort // address and port on which RPC server is listening
			sequencer.chPerPeer[i] = make(chan utils.Message, 10*sequencer.Env.MembersNum)
			go utils.RpcHandler(serviceAddr, "Peer.ReceiveMessage", sequencer.chPerPeer[i], sequencer.Env.Delay)
		}
	})
	// send received message to all peers
	for i := 0; i < sequencer.Env.MembersNum; i++ {
		sequencer.chPerPeer[i] <- arg
	}

	return nil
}
