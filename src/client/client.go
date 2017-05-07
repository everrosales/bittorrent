package btclient

import (
	"btnet"
	"fs"
	// "net"
	"strconv"
	"sync"
	"util"
	"math/rand"
	"time"
)

type status string

const (
	Started   status = "started"
	Completed status = "completed"
	Stopped   status = "stopped"
)

type BTClient struct {
	mu        sync.Mutex
	persister *Persister

	// addr net.Addr
	ip       string
	port     string
	peerId   string
	infoHash string
	status   status

	heartbeatInterval int // number of seconds

	torrentPath string
	torrent     fs.Metadata
	shutdown    chan bool

	numPieces   int
	Pieces      []fs.Piece
	PieceBitmap []bool
	blockBitmap map[int][]bool

	neededPieces chan int
	sendQueue chan btnet.PeerMessage

	// This string is going to be the TCP addr
	peers map[string]btnet.Peer // map from IP to Peer
}

func StartBTClient(ip string, port int, metadataPath string, persister *Persister) *BTClient {

	torrent := fs.ReadTorrent(metadataPath)

	cl := &BTClient{}

	cl.ip = ip
	cl.port = strconv.Itoa(port)
	cl.torrentPath = metadataPath
	cl.torrent = fs.Read(metadataPath) // metadata
	cl.peerId = "-QQ6824-" + util.GenerateRandStr(12)
	cl.infoHash = fs.GetInfoHash(torrent)
	cl.status = Started
	cl.heartbeatInterval = 5

	cl.persister = persister
	cl.shutdown = make(chan bool)

	cl.numPieces = len(cl.torrent.PieceHashes)
	cl.Pieces = make([]fs.Piece, cl.numPieces, cl.numPieces)
	for i, piece := range cl.Pieces {
		piece.Blocks = make([]fs.Block, cl.numBlocks(i), cl.numBlocks(i))
	}
	cl.PieceBitmap = make([]bool, cl.numPieces, cl.numPieces)
	cl.blockBitmap = make(map[int][]bool)
	cl.loadPieces(persister.ReadState())

	cl.neededPieces = make(chan int)
	cl.sendQueue = make(chan btnet.PeerMessage)

	cl.peers = make(map[string]btnet.Peer)

	util.IPrintf("\nClient for %s listening on port %d\n", metadataPath, port)

	go cl.main()
	// cl.listenForPeers()

	return cl
}

func (cl *BTClient) Seed(file string) {
	cl.mu.Lock()
	cl.Pieces = fs.SplitIntoPieces(file, int(cl.torrent.PieceLen))
	for i := range cl.PieceBitmap{
		cl.PieceBitmap[i] = true
	}
	cl.persistPieces()
	cl.mu.Unlock()
}

func (cl *BTClient) Kill() {
	select {
	case _, ok := <-cl.shutdown:
		if ok {
			close(cl.shutdown)
		}
	default:
		// channel already closed
	}
}

// returns true if the client has been ordered to shut down
func (cl *BTClient) CheckShutdown() bool {
	select {
	case _, ok := <-cl.shutdown:
		if !ok {
			return true
		}
	default:
	}
	return false
}

func (cl *BTClient) downloadPiece(piece int) {
	if _, ok := cl.blockBitmap[piece]; !ok {
		cl.blockBitmap[piece] = make([]bool, cl.numBlocks(piece), cl.numBlocks(piece))
	}
	for i:=0; i<cl.numBlocks(piece); i++ {
		go cl.requestBlock(piece, i)
	}
}

func (cl *BTClient) downloadPieces() {
	for {
		if cl.CheckShutdown() {
			return
		}
		piece := <- cl.neededPieces

		cl.mu.Lock()
		if !cl.PieceBitmap[piece] {
			cl.downloadPiece(piece)
		}
		cl.mu.Unlock()

		util.Wait(200)

		cl.mu.Lock()
		if !cl.PieceBitmap[piece] {
			// piece still not downloaded, add it back to queue
			go func() {
				cl.neededPieces <- piece
			}()
		}
		cl.mu.Unlock()
	}
}

func (cl *BTClient) main() {
	go cl.trackerHeartbeat()
	cl.startServer()

	rand.Seed(time.Now().UnixNano())
	go func() {
		for i := range rand.Perm(cl.numPieces) {
			cl.neededPieces <- i
		}
	}()

	go cl.downloadPieces()

	for {
		if cl.CheckShutdown() {
			return
		}
		util.Wait(1000)
	}
}

func (cl *BTClient) numBlocks(piece int) int {
	return fs.NumBlocksInPiece(piece, int(cl.torrent.PieceLen), cl.torrent.GetLength())
}