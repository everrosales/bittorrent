package bttracker

import (
	"sync"
)

type Peer struct {
	peerId     string
	ip         string
	port       int
	uploaded   int
	downloaded int
}

type BTTracker struct {
	file     string
	infoHash string
	mu       sync.Mutex
	peers    []Peer
	count    int
}

func StartBTTracker(file string, port int) *BTTracker {
	tr := &BTTracker{}
	tr.file = file
	// read info_hash from file
	tr.main(port)
	return tr
}
