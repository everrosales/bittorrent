package btclient

//Client

import "sync"
import "net/http"
import "net"
import "time"
import "btnet"
import "fmt"
import "math"
import "crypto/sha1"
import "encoding/base64"
import "encoding/gob"
import "bytes"

import "fs"

type BTClient struct {
	mu        sync.Mutex
	persister *Persister

	// addr net.Addr
	ip   string
	port string

	torrentPath string
	torrent fs.Metadata
	shutdown chan bool

	numPieces int
	numBlocks int
	Pieces []Piece
	PieceBitmap []bool
	blockBitmap map[int][]bool

	peers map[net.Addr]btnet.Peer // map from IP to Peer
}

type Piece struct {
	blocks []Block
}

const BlockSize int = 16384
type Block []byte

func StartBTClient(ip string, port string, metadataPath string, persister *Persister) *BTClient {

	cl := &BTClient{}

	cl.ip = ip
	cl.port = port
	cl.torrentPath = metadataPath
	cl.torrent = fs.Read(metadataPath)

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
		go cl.contactTracker(cl.torrent.TrackerUrl)

		//TODO: only here for compilation
		// fmt.Println(file)

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
	if conn != nil {
	}
	// Send hello message to peer

}

// func (cl *BTClient) listenForPeers() {
//   // Set up stuff
//   cl.startServer()
//
//   //TODO: Send initial hello packets
// }

func (cl *BTClient) startServer() {
	btnet.StartTCPServer(cl.ip+":"+cl.port, cl.messageHandler)
}

func (cl *BTClient) sendBlock(index int, begin int, length int, peer btnet.Peer){
	if !cl.PieceBitmap[index] {
		// we don't have this piece yet
		return
	}
	if length != BlockSize {
		// the requester is using a different block size
		// deny the request for simplicity
		return
	}
	if begin % BlockSize != 0 {
		return
	}
	blockIndex := begin / BlockSize
	data := cl.Pieces[index].blocks[blockIndex]
	go cl.sendPieceMessage(peer, index, begin, length, data)
}

func (cl *BTClient) saveBlock(index int, begin int, length int, block []byte){
	if begin % BlockSize != 0 {
		return
	}
	if length < BlockSize {
		return
	}
	blockIndex := begin / BlockSize
	cl.Pieces[index].blocks[blockIndex] = block[:BlockSize]

	if _, ok := cl.blockBitmap[index]; !ok {
		cl.blockBitmap[index] = make([]bool, cl.numBlocks, cl.numBlocks)
	}
	cl.blockBitmap[index][blockIndex] = true

	if allTrue(cl.blockBitmap[index]) {
		// hash and save piece
		if cl.Pieces[index].hash() != cl.torrent.PieceHashes[index] {
			delete(cl.blockBitmap, index)
			return
		}
		cl.PieceBitmap[index] = true
		cl.persistPieces()
	}
}


func (piece *Piece) hash() string {
	allBytes := []byte{}
	for _, block := range piece.blocks {
		allBytes = append(allBytes, block...)
	}
	hash := make([]byte, 20)
	actualHash := sha1.Sum(allBytes)

	for i := 0; i<20; i++ {
		hash[i] = actualHash[i]
	}
	return base64.URLEncoding.EncodeToString(hash)
}

func (cl *BTClient) persistPieces() {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(cl.Pieces)
	e.Encode(cl.PieceBitmap)
	data := w.Bytes()
	cl.persister.SaveState(data)
}

func (cl *BTClient) loadPieces(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&cl.Pieces)
	d.Decode(&cl.PieceBitmap)
}


func allTrue(arr []bool) bool {
	for _, entry := range arr {
		if !entry {
			return false
		}
	}
	return true
}

func (cl *BTClient) sendPieceMessage(peer btnet.Peer, index int, begin int, length int, data []byte){
	message := btnet.PeerMessage{
		Type: btnet.Piece,
		Index: int32(index),
		Begin: begin,
		Length: length,
		Block: data }
	fmt.Println(message)
	// TODO send message
}

func (cl *BTClient) messageHandler(conn net.Conn) {
	// Process the message
	// Max message size: 2^17 = 131072 (128KB)
	// buf := make([]byte, 131072)
	// bytesRead, err := conn.(*net.TCPConn).Read(buf)
	// fmt.Println(bytesRead)
	// if err != nil {
  //   fmt.Println("hi")
	// 	fmt.Println(err)
	// }
  buf := btnet.ReadMessage(conn.(*net.TCPConn))

	peerMessage := btnet.DecodePeerMessage(buf)
	// Massive switch case that would handle incoming messages depending on message type

	// peerMessage := btnet.PeerMessage{}  // empty for now, TODO
	cl.mu.Lock()
	defer cl.mu.Unlock()

	peer, ok := cl.peers[conn.RemoteAddr()]
  if !ok {
    // InitializePeer
    // TODO: use the actual length len(cl.torrent.PieceHashes)
    cl.peers[conn.RemoteAddr()] = btnet.InitializePeer(conn.RemoteAddr(), 10)
  }

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
		cl.sendBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peer)
	case btnet.Piece:
		cl.saveBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peerMessage.Block)
	case btnet.Cancel:
		// TODO
	default:
		// keepalive
		// TODO?
	}
}
