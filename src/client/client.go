package btclient

import (
	"btnet"
	"fs"
	"math"
	// "net"
	"strconv"
	"sync"
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
	numBlocks   int
	Pieces      []Piece
	PieceBitmap []bool
	blockBitmap map[int][]bool

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
	cl.peerId = util.GenerateRandStr(20)
	cl.infoHash = fs.GetInfoHash(torrent)
	cl.status = Started
	cl.heartbeatInterval = 5

	cl.persister = persister
	cl.shutdown = make(chan bool)

	cl.numPieces = len(cl.torrent.PieceHashes)
	cl.numBlocks = int(math.Ceil(float64(cl.torrent.PieceLen / int64(BlockSize))))
	cl.Pieces = make([]Piece, cl.numPieces, cl.numPieces)
	for _, piece := range cl.Pieces {
		piece.blocks = make([]Block, cl.numBlocks, cl.numBlocks)
	}
	cl.PieceBitmap = make([]bool, cl.numBlocks, cl.numBlocks)
	cl.blockBitmap = make(map[int][]bool)
	cl.loadPieces(persister.ReadState())

	cl.peers = make(map[string]btnet.Peer)

	util.IPrintf("\nClient for %s listening on port %d\n", metadataPath, port)

	go cl.main()
	// cl.listenForPeers()

	return cl
}

func (cl *BTClient) Kill() {
	close(cl.shutdown)
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

func (cl *BTClient) main() {
	go cl.seed()
	cl.startServer()
	for {
		if cl.CheckShutdown() {
			return
		}
		util.Wait(10)
	}
}
