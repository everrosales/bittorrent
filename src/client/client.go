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

// TODO: pruning client's peer list when tracker says that peer is down
// TODO: kill underlying TCP connections when client is killed

const NumDownloaders int = 5
const NumUpdates int = 8

type status string

const (
	Started   status = "started"
	Completed status = "completed"
	Stopped   status = "stopped"
)

type BTClient struct {
	mu        sync.Mutex
	persister *Persister
	alive     bool
	updates   []string

	ip          string
	port        string
	peerId      string
	torrentPath string
	torrentMeta fs.Metadata
	infoHash    string
	outputPath  string

	heartbeatInterval int // number of seconds
	status            status

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
	cl.alive = true
	cl.updates = make([]string, NumUpdates, NumUpdates)

	cl.ip = ip
	cl.port = strconv.Itoa(port)
	cl.peerId = "-QQ6824-" + util.GenerateRandStr(12)
	cl.torrentPath = metadataPath
	cl.torrentMeta = fs.Read(metadataPath) // metadata
	cl.infoHash = fs.GetInfoHash(fs.ReadTorrent(metadataPath))
	cl.outputPath = outputPath

	cl.heartbeatInterval = 1
	cl.status = Started

	cl.numPieces = len(cl.torrentMeta.PieceHashes)
	cl.blockBitmap = make(map[int][]bool)
	cl.neededPieces = make(chan int, 100)
	cl.Pieces = make([]fs.Piece, cl.numPieces, cl.numPieces)
	for i := range cl.Pieces {
		piece := &cl.Pieces[i]
		piece.Blocks = make([]fs.Block, cl.numBlocks(i), cl.numBlocks(i))
	}
	cl.PieceBitmap = make([]bool, cl.numPieces, cl.numPieces)

	cl.peers = make(map[string]*btnet.Peer)

	cl.loadPieces(persister.ReadState())

	util.IPrintf("\nClient for %s listening on port %d\n", metadataPath, port)

	if seedPath != "" {
		cl.Seed(seedPath)
	}
	go cl.main()

	return cl
}

// sends shutdown message
func (cl *BTClient) Kill() {
	cl.lock("killing")
	cl.alive = false
	cl.unlock("killing")
}

// returns true if the client has been ordered to shut down
func (cl *BTClient) CheckShutdown() bool {
	cl.lock("checking shutdown")
	defer cl.unlock("checking shutdown")
	return !cl.alive
}

// returns true if file download is done
func (cl *BTClient) CheckDone() bool {
	cl.lock("checking done")
	if util.AllTrue(cl.PieceBitmap) { // if done, save piece
		if cl.status != Completed {
			pieces := make([]fs.Piece, len(cl.Pieces))
			copy(pieces, cl.Pieces)
			cl.status = Completed
			util.IPrintf("%s: Done downloading, writing to %s\n", cl.port, cl.outputPath)
			cl.unlock("checking done")
			fs.CombinePieces(cl.outputPath, pieces, cl.torrentMeta.Files[0].Length)
		} else {
			cl.unlock("checking done")
		}
		return true
	}
	cl.unlock("checking done")
	return false
}

// check whether download is finished and combines and saves output if it's done
func (cl *BTClient) CheckSaveOutput() {
	for {
		if cl.CheckShutdown() {
			return
		}
		if cl.CheckDone() {
			return
		}
		util.Wait(100)
	}
}

func (cl *BTClient) main() {
	rand.Seed(time.Now().UnixNano())
	go cl.trackerHeartbeat() // start sending heartbeats to tracker
	go cl.startTCPServer()   // start TCP server for communicating with peers

	go func() { // adding the initially needed pieces to the needed queue
		for _, i := range rand.Perm(cl.numPieces) {
			if !cl.atomicGetBitmapElement(i) {
				cl.neededPieces <- i
			}
		}
	}()

	for i := 0; i < NumDownloaders; i++ {
		go cl.downloadPieces()
	}

	if cl.outputPath != "" {
		go cl.CheckSaveOutput()
	}

	for {
		if cl.CheckShutdown() {
			return
		}
		util.Wait(100)
	}
}

func (cl *BTClient) GetStatusString() (string, int) {
	// Here we are dividing by two because of the way the code is
	// currently structured.
	// tl;dr: we dont want to repeat a large chunks of code, so
	// we have a single message handler and peer list for incoming
	// and outgoing messages. Because when dialing a TCP connection
	// opens a new port and binds it to the target peer's listen port,
	// we have two ports for every known peer, the port the peer is
	// listening on and a port that we can use to respond to a requesting
	// peers messages.
	// TODO: Split up the listen messageHandler and the request->response
	//       messageHandler
	cl.lock("status string")
	numPeers := len(cl.peers) / 2
	update := ""
	for _, s := range cl.updates {
		update += s + "\n"
	}
	extraLines := len(cl.updates)
	cl.unlock("status string")
	output := fmt.Sprintf("Known peers: %d\n", numPeers)
	output += "Download status: "
	bitfield, lines := util.BitfieldToString(cl.PieceBitmap, 40)
	output += bitfield + "\n--------\n"
	output += update
	return output, lines + 3 + extraLines
}
