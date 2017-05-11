package btclient

import (
	"btnet"
	"fmt"
	"fs"
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

	ip          string
	port        string
	peerId      string
	torrentPath string
	torrentMeta fs.Metadata
	infoHash    string
	outputPath  string

	heartbeatInterval int // number of seconds
	status            status
	shutdown          chan bool

	numPieces    int
	blockBitmap  map[int][]bool
	neededPieces chan int
	Pieces       []fs.Piece
	PieceBitmap  []bool

	// This string is going to be the TCP addr
	peers map[string]*btnet.Peer // map from IP to Peer
}

func StartBTClient(ip string, port int, metadataPath string, seedPath string, outputPath string, persister *Persister) *BTClient {

	cl := &BTClient{}
	cl.persister = persister

	cl.ip = ip
	cl.port = strconv.Itoa(port)
	cl.peerId = "-QQ6824-" + util.GenerateRandStr(12)
	cl.torrentPath = metadataPath
	cl.torrentMeta = fs.Read(metadataPath) // metadata
	cl.infoHash = fs.GetInfoHash(fs.ReadTorrent(metadataPath))
	cl.outputPath = outputPath

	cl.heartbeatInterval = 5
	cl.status = Started
	cl.shutdown = make(chan bool)

	cl.numPieces = len(cl.torrentMeta.PieceHashes)
	cl.blockBitmap = make(map[int][]bool)
	cl.neededPieces = make(chan int)
	cl.Pieces = make([]fs.Piece, cl.numPieces, cl.numPieces)
	for i := range cl.Pieces {
		piece := &cl.Pieces[i]
		piece.Blocks = make([]fs.Block, cl.numBlocks(i), cl.numBlocks(i))
	}
	cl.PieceBitmap = make([]bool, cl.numPieces, cl.numPieces)

	cl.peers = make(map[string]*btnet.Peer)

	cl.loadPieces(persister.ReadState())

	util.TPrintf("\nClient for %s listening on port %d\n", metadataPath, port)

	if seedPath != "" {
		cl.Seed(seedPath)
	}
	go cl.main()

	return cl
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

func (cl *BTClient) SaveOutput() {
	cl.lock("main/saveoutput")
	pieces := make([]fs.Piece, len(cl.Pieces))
	copy(pieces, cl.Pieces)
	cl.unlock("main/saveoutput")
	fs.CombinePieces(cl.outputPath, pieces, cl.torrentMeta.Files[0].Length)
}

func (cl *BTClient) main() {
	go cl.trackerHeartbeat()
	cl.startServer()

	rand.Seed(time.Now().UnixNano())
	go func() {
		for i := range rand.Perm(cl.numPieces) {
			util.TPrintf("%s: adding piece %d to needed queue\n", cl.port, i)
			cl.neededPieces <- i
		}
	}()

	go cl.downloadPieces()

	if cl.outputPath != "" {
		go func() {
			for {
				cl.lock("main/main")
				if allTrue(cl.PieceBitmap) {
					// all pieces downloaded, save file
					cl.unlock("main/main")
					cl.SaveOutput()
					return
				}
				cl.unlock("main/main")
			}
		}()
	}

	for {
		if cl.CheckShutdown() {
			return
		}
		util.Wait(100)
	}
}

func (cl *BTClient) GetStatusString() (string, int) {
	output := fmt.Sprintf("Known peers: %d\n", len(cl.peers))
	output += "Download status: "
	bitfield, lines := util.BitfieldToString(cl.PieceBitmap, 40)
	output += bitfield + "\n"
	return output, lines + 2
}
