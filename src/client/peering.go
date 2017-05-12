package btclient

import (
	"btnet"
	"bytes"
	"encoding/gob"
	"fs"
	"net"
	"time"
	"util"
)

const peerTimeout = time.Millisecond * 3000
const DialTimeout = time.Millisecond * 100

func (cl *BTClient) startTCPServer() {
	btnet.StartTCPServer(cl.ip+":"+cl.port, cl.messageHandler)
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

func (cl *BTClient) getPeer(addr string) (*btnet.Peer, bool) {
	cl.lock("peering/getPeer")
	p, ok := cl.peers[addr]
	cl.unlock("peering/getPeer")
	return p, ok
}

func (cl *BTClient) getNumPeers() int {
	cl.lock("peering/getNumPeers")
	num := len(cl.peers)
	cl.unlock("peering/getNumPeers")
	return num
}

func (cl *BTClient) requestBlock(piece int, block int) {
	util.TPrintf("%s: want to request block, current peers %v\n", cl.port, cl.peers)
	peerList := cl.getRandomPeerOrder()
	for _, peer := range peerList {
		util.TPrintf("%s: peer bitfield: %v\n", cl.port, peer.Bitfield)
		if peer.Bitfield[piece] && !peer.Status.PeerChoking {
			util.TPrintf("%s: requesting piece %d block %d from peer %s\n", cl.port, piece, block, peer.Addr)
			begin := block * fs.BlockSize
			cl.sendRequestMessage(peer, piece, begin, fs.BlockSize)
		}
	}


}

func (cl *BTClient) sendBlock(index int, begin int, length int, peer *btnet.Peer) {
	util.TPrintf("%s: in sendBlock\n", cl.port)
	if !cl.PieceBitmap[index] {
		util.TPrintf("%s: we don't have this piece\n", cl.port)
		// we don't have this piece yet
		return
	}
	if length != fs.BlockSize {
		util.TPrintf("%s: different block size\n", cl.port)
		// the requester is using a different block size
		// deny the request for simplicity
		return
	}
	if begin%fs.BlockSize != 0 {
		util.TPrintf("%s: not aligned with a block\n", cl.port)
		return
	}
	blockIndex := begin / fs.BlockSize
	util.TPrintf("%s: sending piece %d, block %d\n", cl.port, index, blockIndex)
	data := cl.Pieces[index].Blocks[blockIndex]
	go cl.sendPieceMessage(peer, index, begin, length, data)
}

func (cl *BTClient) saveBlock(index int, begin int, length int, block []byte) {
	util.TPrintf("%s: in saveBlock\n", cl.port)
	if begin%fs.BlockSize != 0 {
		util.TPrintf("%s: not aligned with a block\n", cl.port)
		return
	}
	blockIndex := begin / fs.BlockSize
	util.TPrintf("%s: saving piece %d, block %d\n", cl.port, index, blockIndex)
	cl.lock("peering/saveBlock")
	if length < fs.BlockSize {
		util.TPrintf("%s: len less than BlockSize\n", cl.port)
		cl.Pieces[index].Blocks[blockIndex] = block
	} else {
		cl.Pieces[index].Blocks[blockIndex] = block[:fs.BlockSize]
	}

	if _, ok := cl.blockBitmap[index]; !ok {
		cl.blockBitmap[index] = make([]bool, cl.numBlocks(index), cl.numBlocks(index))
	}
	cl.blockBitmap[index][blockIndex] = true
	util.TPrintf("%s: bitmaps %v\n", cl.port, cl.blockBitmap)

	if allTrue(cl.blockBitmap[index]) {
		util.TPrintf("%s: got all blocks for piece\n", cl.port)
		// hash and save piece
		// TODO: replace this hash
		if cl.Pieces[index].Hash() != cl.torrentMeta.PieceHashes[index] {
			util.TPrintf("%s: hash didn't match\n", cl.port)
			// util.TPrintf("%s: %x != %x\n", cl.port, cl.Pieces[index].Hash(), cl.torrentMeta.PieceHashes[index])
			util.TPrintf("%s: hash lens %d, %d", cl.port, len(cl.Pieces[index].Hash()), len(cl.torrentMeta.PieceHashes[index]))
			delete(cl.blockBitmap, index)
			cl.unlock("peering/saveBlock")
			return
		}
		util.TPrintf("%s: saving piece\n", cl.port)
		cl.PieceBitmap[index] = true
		pieceBitmap := make([]bool, len(cl.PieceBitmap))
		copy(pieceBitmap, cl.PieceBitmap)
		pieces := make([]fs.Piece, len(cl.Pieces))
		copy(pieces, cl.Pieces)
		cl.unlock("peering/saveBlock")
		cl.persister.persistPieces(pieces, pieceBitmap)
		cl.lock("peering/saveBlock")
	}
	cl.unlock("peering/saveBlock")

	for i := range cl.peers {
		// send have message
		cl.sendHaveMessage(cl.peers[i], index, begin, length)
	}
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

func (cl *BTClient) sendRequestMessage(peer *btnet.Peer, index int, begin int, length int) {
	message := btnet.PeerMessage{
		Type:   btnet.Request,
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
	// util.Printf("Got SetupPeerConnections lock\n")
	peer := btnet.InitializePeer(addr, infoHash, peerId, bitfieldLength, conn, cl.PieceBitmap)
	if peer == nil {
		// We got a bad handshake so drop the connection
		return
	}
	cl.lock("peering/setup peer connnections 0")
	cl.peers[addr.String()] = peer
	cl.unlock("peering/setup peer connnections 0")

	// Start go routine that handles the closing of the tcp connection if we dont
	// get a keepAlive signal
	// Separate go routine for sending keepalive signals
	go func() {
		for {
			// util.Printf("Grabbing msgQueue lock\n")
			// util.Printf("Got msgQueue lock\n")
			if peer == nil || peer.Conn.RemoteAddr() == nil {
				// util.Printf("Remote conncetion closed?\n")
				// cl.mu.Unlock()
				return
			}
			var msg btnet.PeerMessage
			select {
			case msg = <-peer.MsgQueue:
				util.TPrintf("Received message from msgqueue - Type: %v\n", msg.Type)
			case <-time.After(peerTimeout / 3):
				msg = btnet.PeerMessage{KeepAlive: true}
			}
			data := btnet.EncodePeerMessage(msg)
			// if (!msg.KeepAlive) {

			util.TPrintf("Sending encoded message from: %v, to: %v, type: %v\n",
				peer.Conn.LocalAddr().String(), peer.Conn.RemoteAddr().String(), msg.Type)
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
				util.WPrintf("err: %v\n", err)
				if peer.Conn.RemoteAddr() != nil {
					util.TPrintf("Closing the connection in peering/SetupPeerConnections\n")
					peer.Conn.Close()
				}
				// cl.mu.Unlock()
				return
			}
			util.TPrintf("Sent message type: %v, to: %s\n", msg.Type, peer.Conn.RemoteAddr().String())
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
					util.TPrintf("Closing the connection in peering keepalive loop\n")
					peer.Conn.Close()
				}
				cl.lock("peering/keepalivetimeout")
				util.WPrintf("KeepAliveTIMEOUT EXCEEDED port: %v\n", cl.port)
				delete(cl.peers, peer.Addr.String())
				cl.unlock("peering/keepalivetimeout")
				return
			}
		}
	}()

	// Start another go routine to read stuff from that channel
	go func() {
		if peer.Conn.RemoteAddr() == nil {
			util.TPrintf("Connection Closed!\n")
		} else {
			util.TPrintf("Connection Alive!\n")
		}
		cl.messageHandler(&peer.Conn)
	}()
}

func (cl *BTClient) SendPeerMessage(addr *net.TCPAddr, message btnet.PeerMessage) {
	cl.lock("peering/sendPeerMessage")
	peer, ok := cl.peers[addr.String()]
	if !ok {
		cl.unlock("peering/sendPeerMessage")
		cl.SetupPeerConnections(addr, nil)
		peer, ok = cl.peers[addr.String()]
		if !ok {
			util.WPrintf("Failed to establish a connection\n")
			return
		}
		peer.MsgQueue <- message
		return
	}

	cl.unlock("peering/sendPeerMessage")
	peer.MsgQueue <- message
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
	util.TPrintf("~~~ Got a connection! ~~~\n")
	for ok {
		// Process the message
		buf, err := btnet.ReadMessage(conn)
		if err != nil {
			util.WPrintf("%s\n", err)
			conn.Close()
			return
		}
		peerMessage := btnet.DecodePeerMessage(buf, len(cl.torrentMeta.PieceHashes))
		util.TPrintf("Received PeerMessage, type: %v, from: %s\n", peerMessage.Type, conn.RemoteAddr().String())
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

		if !peerMessage.KeepAlive {
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
				util.TPrintf("%s: received request msg\n", cl.port)
				cl.sendBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peer)
			case btnet.Piece:
				util.TPrintf("%s: received piece msg\n", cl.port)
				cl.saveBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peerMessage.Block)
			case btnet.Cancel:
				// TODO
			default:
				// keepalive
				// TODO?
			}
		}
		// Update okay to make sure that we still have a connection
		// _, ok = cl.peers[conn.RemoteAddr().String()]
		ok = conn.RemoteAddr() != nil
	}
	util.TPrintf("%s: exiting messageHandler\n", cl.port)
}
