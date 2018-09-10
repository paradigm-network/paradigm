package sequentia

import (
	"math/rand"

	"github.com/paradigm-network/paradigm/network/peer"
)

type PeerSelector interface {
	Peers() []peer.Peer
	UpdateLast(peer string)
	Next() peer.Peer
}

//+++++++++++++++++++++++++++++++++++++++
//RANDOM

type RandomPeerSelector struct {
	peers []peer.Peer
	last  string
}

func NewRandomPeerSelector(participants []peer.Peer, localAddr string) *RandomPeerSelector {
	_, peers := peer.ExcludePeer(participants, localAddr)
	return &RandomPeerSelector{
		peers: peers,
	}
}

func (ps *RandomPeerSelector) Peers() []peer.Peer {
	return ps.peers
}

func (ps *RandomPeerSelector) UpdateLast(peer string) {
	ps.last = peer
}

func (ps *RandomPeerSelector) Next() peer.Peer {
	selectablePeers := ps.peers
	if len(selectablePeers) > 1 {
		_, selectablePeers = peer.ExcludePeer(selectablePeers, ps.last)
	}
	i := rand.Intn(len(selectablePeers))
	peer := selectablePeers[i]
	return peer
}
