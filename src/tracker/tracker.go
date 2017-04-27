package bttracker

import (
	"sync"
)

type peerStatus string

const (
	Started   peerStatus = "started"
	Completed peerStatus = "completed"
	Stopped   peerStatus = "stopped"
	Empty     peerStatus = "empty"
)

type Peer struct {
	peerId     string
	ip         string
	port       int
	uploaded   int
	downloaded int
	left       int
	status     peerStatus
}

type BTTracker struct {
	file     string
	infoHash string
	mu       sync.Mutex
	peers    map[string]Peer
}

func StartBTTracker(file string, port int) *BTTracker {
	tr := &BTTracker{}
	tr.file = file
	tr.peers = make(map[string]Peer)
	// TODO: read info_hash from file
	tr.main(port)
	return tr
}
