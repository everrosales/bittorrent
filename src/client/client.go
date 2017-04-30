package btclient

//Client

import "sync"
import "net/http"
import "net"
import "time"
import "btnet"
import "fmt"

type TorrentMetadata struct {
	path string
}

type BTClient struct {
	mu        sync.Mutex
	persister *Persister

	// addr net.Addr
  ip string
  port string

	files    map[TorrentMetadata]string // map from torrent metadata paths to their local download paths
	seeding  []TorrentMetadata          // List of previous Torrent files and their Metadata
	shutdown chan bool

	peers map[net.Addr]btnet.Peer // map from IP to Peer
}

func StartBTClient(ip string, port string, persister *Persister) *BTClient {
	cl := &BTClient{}

	cl.ip = ip
	cl.port = port

	cl.persister = persister
	cl.files = make(map[TorrentMetadata]string)
	cl.seeding = []TorrentMetadata{}
	cl.shutdown = make(chan bool)

	go cl.main()
	// cl.listenForPeers()

	return cl
}

func (cl *BTClient) dialTimeout() time.Duration {
	return time.Millisecond * 100
}

func (cl *BTClient) Kill() {
	close(cl.shutdown)
}

func (cl *BTClient) contactTracker(url string) {
	http.Get(url)
}

func (cl *BTClient) seed() {
	for {
		if cl.checkShutdown() {
			return
		}
		cl.mu.Lock()
		for _, file := range cl.seeding {
			url := "" // TODO get from file
			go cl.contactTracker(url)

      //TODO: only here for compilation
      fmt.Println(file)
		}

    cl.mu.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (cl *BTClient) main() {
	go cl.seed()
	cl.startServer()
	for {
		if cl.checkShutdown() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}


// returns true if the client has been ordered to shut down
func (cl *BTClient) checkShutdown() bool {
	select {
	case _, ok := <-cl.shutdown:
		if !ok {
			return true
		}
	default:
	}
	return false
}

func (cl *BTClient) connectToPeer(addr string) {
	conn, err := net.DialTimeout("tcp", addr, cl.dialTimeout())
	if err != nil {
		// TODO: try again or mark peer as down
	}
	// Create hello message
  // TODO: here for compilation hehe
  if conn!=nil {}
	// Send hello message to peer

}

// func (cl *BTClient) listenForPeers() {
//   // Set up stuff
//   cl.startServer()
//
//   //TODO: Send initial hello packets
// }

func (cl *BTClient) startServer() {
	btnet.StartTCPServer(cl.ip + ":" + cl.port, cl.messageHandler)
}

func (cl *BTClient) sendPiece(index int, begin int, length int, peer btnet.Peer){
	// TODO
}

func (cl *BTClient) savePiece(index int, begin int, length int, piece []byte){
	// TODO
}

func (cl *BTClient) messageHandler(conn net.Conn) {
	// Process the message
	// Max message size: 2^17 = 131072 (128KB)
	buf := make([]byte, 131072)
	bytesRead, err := conn.(*net.TCPConn).Read(buf)
	fmt.Println(bytesRead)
	if err != nil {
    fmt.Println("hi")
		fmt.Println(err)
	}
	peerMessage := btnet.DecodePeerMessage(buf)
	// Massive switch case that would handle incoming messages depending on message type

	// peerMessage := btnet.PeerMessage{}  // empty for now, TODO
	cl.mu.Lock()
	defer cl.mu.Unlock()

	peer := cl.peers[conn.RemoteAddr()]

	switch peerMessage.Type {
	case btnet.Choke:
		peer.Status.PeerChoking = true
	case btnet.Unchoke:
		peer.Status.PeerChoking = false
	case btnet.Interested:
		peer.Status.PeerInterested = true
	case btnet.NotInterested:
		peer.Status.PeerInterested = false
	case btnet.Have:
		peer.Bitfield[peerMessage.Index] = true
	case btnet.Bitfield:
		peer.Bitfield = peerMessage.Bitfield
	case btnet.Request:
		cl.sendPiece(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peer)
	case btnet.Piece:
		cl.savePiece(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peerMessage.Block)
	case btnet.Cancel:
		// TODO
	default:
		// keepalive
		// TODO?
	}
}
