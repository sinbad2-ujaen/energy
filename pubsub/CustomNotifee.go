package pubsub

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
)

type CustomNotifee struct{}

func (n *CustomNotifee) HandlePeerFound(info peer.AddrInfo) {
	fmt.Printf("Found a new peer: %s\n", info.ID.String())
}

func (n *CustomNotifee) HandlePeerLost(info peer.AddrInfo) {
	fmt.Printf("Lost a peer: %s\n", info.ID.String())
}
