package btclient

import (
	"btnet"
	"fs"
	// "net"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"util"
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

	// This string is going to be the TCP addr
	peers map[string]*btnet.Peer // map from IP to Peer
}

func StartBTClient(ip string, port int, metadataPath string, seedPath string, persister *Persister) *BTClient {

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
	for i := range cl.Pieces {
		piece := &cl.Pieces[i]
		piece.Blocks = make([]fs.Block, cl.numBlocks(i), cl.numBlocks(i))
	}
	cl.PieceBitmap = make([]bool, cl.numPieces, cl.numPieces)
	cl.blockBitmap = make(map[int][]bool)
	cl.loadPieces(persister.ReadState())

	cl.neededPieces = make(chan int)

	cl.peers = make(map[string]*btnet.Peer)

	util.IPrintf("\nClient for %s listening on port %d\n", metadataPath, port)

	if seedPath != "" {
		cl.Seed(seedPath)
	}
	go cl.main()

	return cl
}

func (cl *BTClient) Seed(file string) {
	// util.Printf("Grabbing Seed lock\n")
	cl.mu.Lock()
	// util.Printf("Got seed lock\n")
	cl.Pieces = fs.SplitIntoPieces(file, int(cl.torrent.PieceLen))
	for i := range cl.PieceBitmap {
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
	util.TPrintf("need to download piece %d\n", piece)
	if _, ok := cl.blockBitmap[piece]; !ok {
		cl.blockBitmap[piece] = make([]bool, cl.numBlocks(piece), cl.numBlocks(piece))
	}
	for i := 0; i < cl.numBlocks(piece); i++ {
		// go cl.requestBlock(piece, i)
		cl.requestBlock(piece, i)
	}
}

func (cl *BTClient) downloadPieces() {
	for {
		if cl.CheckShutdown() {
			return
		}
		piece := <-cl.neededPieces
		// util.Printf("Grabbing downloadPieces lock\n")
		cl.mu.Lock()
		// util.Printf("Got downloadPieces lock\n")
		if !cl.PieceBitmap[piece] {
			cl.mu.Unlock()
			cl.downloadPiece(piece)
			cl.mu.Lock()
		}
		cl.mu.Unlock()

		util.Wait(200)

		// util.Printf("Grabbing downloadPieces lock v2\n")
		cl.mu.Lock()
		// util.Printf("Got downloadPieces lock v2\n")
		if !cl.PieceBitmap[piece] {
			util.TPrintf("piece was not downloaded")
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
			util.TPrintf("adding piece %d to needed queue\n", i)
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
