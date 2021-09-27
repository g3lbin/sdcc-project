package rpcpeer

import (
	"github.com/sdcc-project/internal/pkg/utils"
)

type Peer struct {

}

var ChFromPeers chan utils.Sender

func (registry *Peer) ReceiveMessage(arg utils.Sender, res *int) error {
	ChFromPeers <- arg
	*res = 1

	return nil
}
