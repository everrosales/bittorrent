package bttracker

import (
	"fs"
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

// Instantiate a new BTTracker
func StartBTTracker(path string, port int) {
	tr := &BTTracker{}
	tr.file = path
	tr.peers = make(map[string]Peer)
	torrent := fs.ReadTorrent(path)
	tr.infoHash = fs.GetInfoHash(torrent)
	tr.main(port)
	return
}
