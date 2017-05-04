package bttracker

import (
	"fs"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
	"util"
)

type peerId string
type peerStatus string

const (
	Started   peerStatus = "started"
	Completed peerStatus = "completed"
	Stopped   peerStatus = "stopped"
	Empty     peerStatus = "empty"
)

const MaxPeers = 50
const PeerWaitTime = time.Duration(10) * time.Second

// private tracker's peer state
type peer struct {
	peerId     peerId
	ip         string
	port       int
	uploaded   int
	downloaded int
	left       int
	status     peerStatus
	lastSeen   time.Time
}

// tracker state
type BTTracker struct {
	file     string
	infoHash string
	mu       sync.Mutex
	peers    map[peerId]peer
	port     int
	shutdown chan bool
	srv      *http.Server
}

// Instantiate a new BTTracker
func StartBTTracker(path string, port int) *BTTracker {
	tr := &BTTracker{}
	tr.file = path
	tr.port = port
	tr.peers = make(map[peerId]peer)
	tr.shutdown = make(chan bool)
	torrent := fs.ReadTorrent(path)
	tr.infoHash = fs.GetInfoHash(torrent)

	util.IPrintf("Tracker for %s listening on port %d - escaped infohash %s\n", tr.file, port, url.QueryEscape(tr.infoHash))
	go tr.main(port)
	go tr.watchPeers()
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

func (tr *BTTracker) getPeerList() []map[string]string {
	peerList := [](map[string]string){}
	count := 0

	for _, v := range tr.peers {
		p := map[string]string{"peer id": string(v.peerId), "ip": v.ip, "port": strconv.Itoa(v.port)}
		peerList = append(peerList, p)
		count += 1
		if count == MaxPeers {
			return peerList
		}
	}
	return peerList
}

// TODO: set timeout for peers in list
func (tr *BTTracker) watchPeers() {
	for {
		timeNow := time.Now()
		tr.mu.Lock()
		for k, v := range tr.peers {
			if timeNow.After((v.lastSeen).Add(PeerWaitTime)) {
				delete(tr.peers, k)
			}
		}
		tr.mu.Unlock()
		util.Wait(1000)
	}
}
