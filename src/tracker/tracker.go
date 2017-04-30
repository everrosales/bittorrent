package bttracker

import (
	"fs"
	"strconv"
	"sync"
)

type peerStatus string

const (
	Started   peerStatus = "started"
	Completed peerStatus = "completed"
	Stopped   peerStatus = "stopped"
	Empty     peerStatus = "empty"
)

// private tracker's peer state
type peer struct {
	peerId     string
	ip         string
	port       int
	uploaded   int
	downloaded int
	left       int
	status     peerStatus
}

// tracker state
type BTTracker struct {
	file     string
	infoHash string
	mu       sync.Mutex
	peers    map[string]peer
}

// Instantiate a new BTTracker
func StartBTTracker(path string, port int) {
	tr := &BTTracker{}
	tr.file = path
	tr.peers = make(map[string]peer)
	torrent := fs.ReadTorrent(path)
	tr.infoHash = fs.GetInfoHash(torrent)
	tr.main(port)
	return
}

func (tr BTTracker) getPeers() []map[string]string {
	peers := [](map[string]string){}
	for _, v := range tr.peers {
		p := map[string]string{"peer id": v.peerId, "ip": v.ip, "port": strconv.Itoa(v.port)}
		peers = append(peers, p)
	}
	return peers
}

// TODO: set timeout for peers in list
