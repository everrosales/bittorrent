package btclient

import (
	"btnet"
	"bytes"
	"encoding/gob"
	"errors"
	"fs"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"util"
)

const peerTimeout = time.Millisecond * 3000
const DialTimeout = time.Millisecond * 100

func (cl *BTClient) startServer() {
	btnet.StartTCPServer(cl.ip+":"+cl.port, cl.messageHandler)
}

type trackerReq struct {
	peerId     string
	ip         string
	port       string
	uploaded   int
	downloaded int
	left       int
	infoHash   string
	status     status
}

type TrackerRes struct {
	Interval int                 `bencode:"interval"`
	Peers    []map[string]string `bencode:"peers"`
	Failure  string              `bencode:"failure reason"`
}

func sendRequest(addr string, req *trackerReq) ([]byte, error) {
	url := addr + "/?peer_id=" + url.QueryEscape(req.peerId) +
		"&port=" + req.port + "&ip=" + req.ip + "&uploaded=" +
		strconv.Itoa(req.uploaded) + "&downloaded=" + strconv.Itoa(req.downloaded) +
		"&left=" + strconv.Itoa(req.left) + "&info_hash=" +
		url.QueryEscape(req.infoHash) + "&status=" + string(req.status)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Error sending request")
	}
	if resp.Status != "200 OK" {
		return nil, errors.New("Wrong response status code")
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, errors.New("Failure reading response body")
	}
	return bodyBytes, nil
}

func (cl *BTClient) contactTracker(baseUrl string) TrackerRes {
	// TODO: update uploaded, downloaded, and left
	// util.Printf("Grabbing contactTracker lock\n")
	cl.mu.Lock()
	// util.Printf("Got contactTracker lock\n")
	request := trackerReq{cl.peerId, cl.ip, cl.port, 0, 0, 0, cl.infoHash, cl.status}
	cl.mu.Unlock()
	byteRes, err := sendRequest(baseUrl, &request)
	if err != nil {
		util.EPrintf("Received error sending to tracker: %s\n", err)
	}
	res := TrackerRes{}
	fs.Decode(string(byteRes), &res)
	if res.Failure != "" {
		util.EPrintf("Received error from tracker: %s\n", res.Failure)
	}
	util.IPrintf("Contacting tracker at %s (%d peers)\n", baseUrl, len(res.Peers))
	// util.Printf("Grabbing contactTracker lock v2\n")
	cl.mu.Lock()
	// util.Printf("Got contactTracker lock v2\n")
	cl.heartbeatInterval = res.Interval
	cl.mu.Unlock()
	return res
}

// func (cl *BTClient) listenForPeers() {
//   // Set up stuff
//   cl.startServer()
//
//   //TODO: Send initial hello packets
// }

func (cl *BTClient) trackerHeartbeat() {
	for {
		if cl.CheckShutdown() {
			return
		}
		res := cl.contactTracker(cl.torrent.TrackerUrl)
		for _, p := range res.Peers {
			util.TPrintf("peerId %s, ip %s, port %s\n", p["peer id"], p["ip"], p["port"])
			addr, err := net.ResolveTCPAddr("tcp", p["ip"]+":"+p["port"])
			if err != nil {
				panic(err)
			}
			myAddr, err := net.ResolveTCPAddr("tcp", cl.ip+":"+cl.port)
			if addr.String() != myAddr.String() {
				util.TPrintf("sending initial message to %v", addr)
				cl.SendPeerMessage(addr, btnet.PeerMessage{KeepAlive: true})
			}
		}
		// util.Printf("Grabbing trackerHeartbeat lock\n")
		cl.mu.Lock()
		// util.Printf("Got trackerHeartbeat lock\n")
		wait := cl.heartbeatInterval * 1000
		cl.mu.Unlock()
		util.Wait(wait)
	}
}

func (cl *BTClient) connectToPeer(addr string) {
	conn, err := net.DialTimeout("tcp", addr, DialTimeout)
	if err != nil {
		// TODO: try again or mark peer as down
	}
	// Create hello message
	// TODO: here for compilation hehe
	if conn != nil {
	}
	// Send hello message to peer
}

func (cl *BTClient) requestBlock(piece int, block int) {
	// util.Printf("Grabbing requestBlock lock\n")
	// cl.mu.Lock()
	// util.Printf("Got requestBlock lock\n")
	for addr, peer := range cl.peers {
		// util.TPrintf("peer bitfield: %v\n", peer.Bitfield)
		if peer.Bitfield[piece] && !peer.Status.PeerChoking {
			util.Printf("requesting piece %d block %d from peer %s\n", piece, block, addr)
			begin := block * fs.BlockSize
			cl.sendRequestMessage(peer, piece, begin, fs.BlockSize)
		}
	}

	// cl.mu.Unlock()
}

func (cl *BTClient) sendBlock(index int, begin int, length int, peer *btnet.Peer) {
	if !cl.PieceBitmap[index] {
		// we don't have this piece yet
		return
	}
	if length != fs.BlockSize {
		// the requester is using a different block size
		// deny the request for simplicity
		return
	}
	if begin%fs.BlockSize != 0 {
		return
	}
	blockIndex := begin / fs.BlockSize
	util.TPrintf("sending piece %d, block %d", index, blockIndex)
	data := cl.Pieces[index].Blocks[blockIndex]
	go cl.sendPieceMessage(peer, index, begin, length, data)
}

func (cl *BTClient) saveBlock(index int, begin int, length int, block []byte) {
	if begin%fs.BlockSize != 0 {
		return
	}
	if length < fs.BlockSize {
		return
	}
	blockIndex := begin / fs.BlockSize
	util.TPrintf("saving piece %d, block %d", index, blockIndex)
	cl.Pieces[index].Blocks[blockIndex] = block[:fs.BlockSize]

	if _, ok := cl.blockBitmap[index]; !ok {
		cl.blockBitmap[index] = make([]bool, cl.numBlocks(index), cl.numBlocks(index))
	}
	cl.blockBitmap[index][blockIndex] = true

	if allTrue(cl.blockBitmap[index]) {
		// hash and save piece
		if cl.Pieces[index].Hash() != cl.torrent.PieceHashes[index] {
			delete(cl.blockBitmap, index)
			return
		}
		cl.PieceBitmap[index] = true
		cl.persistPieces()
	}

	for _, peer := range cl.peers {
		// send have message
		cl.sendHaveMessage(peer, index, begin, length)
	}
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

// Helper
func allTrue(arr []bool) bool {
	for _, entry := range arr {
		if !entry {
			return false
		}
	}
	return true
}

func (cl *BTClient) sendRequestMessage(peer *btnet.Peer, index int, begin int, length int) {
	message := btnet.PeerMessage{
		Type:   btnet.Piece,
		Index:  int32(index),
		Begin:  begin,
		Length: length}
	// util.TPrintf("sending message - %v\n", message)
	cl.SendPeerMessage(&peer.Addr, message)
}

func (cl *BTClient) sendPieceMessage(peer *btnet.Peer, index int, begin int, length int, data []byte) {
	message := btnet.PeerMessage{
		Type:   btnet.Piece,
		Index:  int32(index),
		Begin:  begin,
		Length: length,
		Block:  data}
	// util.TPrintf("sending message - %v\n", message)
	cl.SendPeerMessage(&peer.Addr, message)
}

func (cl *BTClient) sendHaveMessage(peer *btnet.Peer, index int, begin int, length int) {
	message := btnet.PeerMessage{
		Type:   btnet.Have,
		Index:  int32(index),
		Begin:  begin,
		Length: length}
	// util.TPrintf("sending message - %v\n", message)
	cl.SendPeerMessage(&peer.Addr, message)
}

func (cl *BTClient) SetupPeerConnections(addr *net.TCPAddr, conn *net.TCPConn) {
	// Try dialing
	// connection := DoDial(addr, data)

	infoHash := fs.GetInfoHash(fs.ReadTorrent(cl.torrentPath))
	peerId := cl.peerId
	bitfieldLength := cl.numPieces
	// util.Printf("Grabbing SetupPeerConnections lock\n")
	cl.mu.Lock()
	// util.Printf("Got SetupPeerConnections lock\n")
	peer := btnet.InitializePeer(addr, infoHash, peerId, bitfieldLength, conn, cl.PieceBitmap)
	if peer == nil {
		// We got a bad handshake so drop the connection
		cl.mu.Unlock()
		return
	}
	// cl.mu.Lock()
	cl.peers[addr.String()] = peer
	cl.mu.Unlock()

	// Start go routine that handles the closing of the tcp connection if we dont
	// get a keepAlive signal
	// Separate go routine for sending keepalive signals
	go func() {
		for {
			// util.Printf("Grabbing msgQueue lock\n")
			cl.mu.Lock()
			// util.Printf("Got msgQueue lock\n")
			if peer == nil || peer.Conn.RemoteAddr() == nil {
				// util.Printf("Remote conncetion closed?\n")
				cl.mu.Unlock()
				return
			}
			var msg btnet.PeerMessage
			select {
			case msg = <-peer.MsgQueue:
				// util.TPrintf("Received message from msgqueue - Type: %v\n", msg.Type)
			case <-time.After(peerTimeout / 3):
				msg = btnet.PeerMessage{KeepAlive: true}
			}
			data := btnet.EncodePeerMessage(msg)
			// if (!msg.KeepAlive) {

			util.TPrintf("Sending encoded message from: %v, to: %v, type: %v, data: %v\n",
				peer.Conn.LocalAddr().String(), peer.Conn.RemoteAddr().String(), msg.Type, data)
			// }
			// We dont need a lock if only this thread is sending out TCP messages
			// if (peer.Conn == nil) {
			//     util.EPrintf("FFS\n")
			// }

			// util.Printf("peer.Conn.RemoteAddr(): %v\n", peer.Conn.RemoteAddr())
			// peer.Conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500))
			_, err := peer.Conn.Write(data)
			if err != nil {
				// Connection is probably closed
				// TODO: Not sure if this is the right way of checking this
				util.EPrintf("err: %v\n", err)
				if peer.Conn.RemoteAddr() != nil {
					// util.EPrintf("Closing the connection\n")
					peer.Conn.Close()
				}
				cl.mu.Unlock()
				return
			}
			util.Printf("Sent message type: %v\n", msg.Type)
			cl.mu.Unlock()
		}
	}()

	// KeepAlive loop
	go func() {
		for {
			select {
			case <-peer.KeepAlive:
				// Do nothing, this is good
			case <-time.After(peerTimeout):
				if peer.Conn.RemoteAddr() != nil {
					peer.Conn.Close()
					util.EPrintf("Closing the connection again\n")
				}
				// util.Printf("Grabbing KeepAliveTimeout lock\n")
				cl.mu.Lock()
				// util.Printf("Got KeepAliveTimeout lock\n")
				util.EPrintf("KeepAliveTIMEOUT EXCEEDED port: %v\n", cl.port)
				delete(cl.peers, peer.Addr.String())
				// cl.peers[addr] = nil
				cl.mu.Unlock()
				return
			}
		}
	}()

	// Start another go routine to read stuff from that channel
	go func() {
		if peer.Conn.RemoteAddr() == nil {
			util.Printf("Connection Closed!\n")
		} else {
			util.Printf("Connection Alive!\n")
		}
		cl.messageHandler(&peer.Conn)
	}()
}

func (cl *BTClient) SendPeerMessage(addr *net.TCPAddr, message btnet.PeerMessage) {
	cl.mu.Lock()
	peer, ok := cl.peers[addr.String()]
	if !ok {
		cl.mu.Unlock()
		cl.SetupPeerConnections(addr, nil)
		cl.mu.Lock()
		peer, _ = cl.peers[addr.String()]
	}
	peer.MsgQueue <- message
	cl.mu.Unlock()
}

func (cl *BTClient) messageHandler(conn *net.TCPConn) {
	// Max message size: 2^17 = 131072 (128KB)
	// buf := make([]byte, 131072)
	// bytesRead, err := conn.(*net.TCPConn).Read(buf)
	// fmt.Println(bytesRead)
	// if err != nil {
	//   fmt.Println("hi")
	// 	fmt.Println(err)
	// }
	// Check if this is a new connection
	// If so we need to initialize the Peer
	util.TPrintf("~~~ Got a connection! ~~~\n")
	if conn.RemoteAddr() == nil {
		return
	}
	peer, ok := cl.peers[conn.RemoteAddr().String()]
	if !ok {
		// This internally calls messageHandler in a separate goRoutine
		cl.SetupPeerConnections(conn.RemoteAddr().(*net.TCPAddr), conn)
		return
	}
	// peer, ok = cl.peers[conn.RemoteAddr().String()]

	for ok {
		// Process the message
		buf, err := btnet.ReadMessage(conn)
		if err != nil {
			util.EPrintf("%s\n", err)
			return
		}
		peerMessage := btnet.DecodePeerMessage(buf)
		util.TPrintf("Received PeerMessage, type: %v\n%v\n", peerMessage.Type, peerMessage)
		// Massive switch case that would handle incoming messages depending on message type

		// peerMessage := btnet.PeerMessage{}  // empty for now, TODO
		// util.TPrintf("Grabbing lock\n")
		// cl.mu.Lock()
		// defer cl.mu.Unlock()
		// util.TPrintf("Got Lock\n")

		// if peerMessage.KeepAlive {
		// peer.KeepAlive <- true
		// }

		// Really anytime we receive a message we should treat this as
		// a KeepAlive message
		peer.KeepAlive <- true

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
		// Update okay to make sure that we still have a connection
		// _, ok = cl.peers[conn.RemoteAddr().String()]
		ok = conn.RemoteAddr() != nil
	}
	util.Printf("\nExiting messageHandler\n\n")
}
