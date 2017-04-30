package bttracker

import (
	"fs"
	"net/http"
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
	shutdown chan bool
	srv      *http.Server
}

// Instantiate a new BTTracker
func StartBTTracker(path string, port int) *BTTracker {
	tr := &BTTracker{}
	tr.file = path
	tr.peers = make(map[string]peer)
	tr.shutdown = make(chan bool)
	torrent := fs.ReadTorrent(path)
	tr.infoHash = fs.GetInfoHash(torrent)
	go tr.main(port)
	return tr
}

func (tr *BTTracker) Kill() {
	close(tr.shutdown)
}

// returns true if the tracker has been ordered to shut down
func (tr *BTTracker) CheckShutdown() bool {
	select {
	case _, ok := <-tr.shutdown:
		if !ok {
			return true
		}
	default:
	}
	return false
}

func (tr *BTTracker) getPeers() []map[string]string {
	peers := [](map[string]string){}
	for _, v := range tr.peers {
		p := map[string]string{"peer id": v.peerId, "ip": v.ip, "port": strconv.Itoa(v.port)}
		peers = append(peers, p)
	}
	return peers
}

// TODO: set timeout for peers in list
