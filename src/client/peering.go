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

func (cl *BTClient) requestBlock(piece int, block int) {
	cl.lock("peering/requestBlock")
	util.TPrintf("%s: want to request piece %d block %d, current peers %v\n", cl.port, piece, block, cl.peers)
	peerList := cl.getRandomPeerOrder()
	port := cl.port
	cl.unlock("peering/requestBlock")

	for _, peer := range peerList {
		if peer.GetBitfield()[piece] && !peer.GetStatus().PeerChoking {
			util.TPrintf("%s: requesting piece %d block %d from peer %s\n", port, piece, block, peer.Addr)
			begin := block * fs.BlockSize
			cl.sendRequestMessage(peer, piece, begin, fs.BlockSize)
		}
	}

}

func (cl *BTClient) sendBlock(index int, begin int, length int, peer *btnet.Peer) {
	if !cl.PieceBitmap[index] {
		util.TPrintf("%s: we don't have this piece\n", cl.port)
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

	if allTrue(cl.blockBitmap[index]) {
		// hash and save piece
		if cl.Pieces[index].Hash() != cl.torrentMeta.PieceHashes[index] {
			util.WPrintf("%s: hashes didn't match - lengths: %d, %d", cl.port, len(cl.Pieces[index].Hash()), len(cl.torrentMeta.PieceHashes[index]))
			delete(cl.blockBitmap, index)
			cl.unlock("peering/saveBlock")
			return
		}
		util.TPrintf("%s: saving piece %d\n", cl.port, index)
		cl.PieceBitmap[index] = true
		pieceBitmap := make([]bool, len(cl.PieceBitmap))
		copy(pieceBitmap, cl.PieceBitmap)
		pieces := make([]fs.Piece, len(cl.Pieces))
		copy(pieces, cl.Pieces)
		cl.persister.persistPieces(pieces, pieceBitmap)
	}
	cl.unlock("peering/saveBlock")

	for _, addr := range cl.atomicGetPeerAddrs() {
		// send have message
		p, ok := cl.atomicGetPeer(addr)
		if !ok {
			util.EPrintf("%s: peer %s not found\n", cl.port, addr)
		} else {
			cl.sendHaveMessage(p, index, begin, length)
		}
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
	cl.SendPeerMessage(&peer.Addr, message)
}

func (cl *BTClient) sendPieceMessage(peer *btnet.Peer, index int, begin int, length int, data []byte) {
	message := btnet.PeerMessage{
		Type:   btnet.Piece,
		Index:  int32(index),
		Begin:  begin,
		Length: length,
		Block:  data}
	cl.SendPeerMessage(&peer.Addr, message)
}

func (cl *BTClient) sendHaveMessage(peer *btnet.Peer, index int, begin int, length int) {
	message := btnet.PeerMessage{
		Type:   btnet.Have,
		Index:  int32(index),
		Begin:  begin,
		Length: length}
	cl.SendPeerMessage(&peer.Addr, message)
}

func (cl *BTClient) SetupPeerConnections(addr *net.TCPAddr, conn *net.TCPConn) {
	// Try dialing
	// connection := DoDial(addr, data)

	infoHash := fs.GetInfoHash(fs.ReadTorrent(cl.torrentPath))
	peerId := cl.peerId
	bitfieldLength := cl.numPieces
	peer := btnet.InitializePeer(addr, infoHash, peerId, bitfieldLength, conn, cl.PieceBitmap)
	if peer == nil {
		// We got a bad handshake so drop the connection
		return
	}
	cl.atomicSetPeer(addr.String(), peer)

	// Start go routine that handles the closing of the tcp connection if we dont
	// get a keepAlive signal
	// Separate go routine for sending keepalive signals
	go func() {
		for {
			if peer == nil || peer.Conn.RemoteAddr() == nil {
				return
			}
			var msg btnet.PeerMessage
			select {
			case msg = <-peer.MsgQueue:
				// newMessage = cl.checkIfPending(msg)
				util.TPrintf("Received message from msgqueue - Type: %v\n", msg.Type)
			case <-time.After(peerTimeout / 3):
				msg = btnet.PeerMessage{KeepAlive: true}
			}

			data := btnet.EncodePeerMessage(msg)
			util.TPrintf("Sending encoded message from: %v, to: %v, type: %v\n",
				peer.Conn.LocalAddr().String(), peer.Conn.RemoteAddr().String(), msg.Type)

			_, err := peer.Conn.Write(data)
			peer.MarkMessageSent(msg)
			if err != nil {
				// Connection is probably closed
				// TODO: Not sure if this is the right way of checking this
				util.WPrintf("err: %v\n", err)
				if peer.Conn.RemoteAddr() != nil {
					peer.Conn.Close()
				}
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
					peer.Conn.Close()
				}
				cl.atomicDeletePeer(peer.Addr.String())
				return
			}
		}
	}()

	// Start another go routine to read stuff from that channel
	go func() {
		if peer.Conn.RemoteAddr() == nil {
			util.TPrintf("Connection closed!\n")
		} else {
			util.TPrintf("Connection alive!\n")
		}
		cl.messageHandler(&peer.Conn)
	}()
}

func (cl *BTClient) SendPeerMessage(addr *net.TCPAddr, message btnet.PeerMessage) {
	peer, ok := cl.atomicGetPeer(addr.String())
	if !ok {
		cl.SetupPeerConnections(addr, nil)
		peer, ok = cl.atomicGetPeer(addr.String())
		if !ok {
			util.WPrintf("%d: failed to establish a connection with %s\n", cl.port, addr)
			return
		}
	}

	// peer.MsgQueue <- message
	peer.AddToMessageQueue(message)
	return
}

func (cl *BTClient) messageHandler(conn *net.TCPConn) {
	// Check if this is a new connection
	// If so we need to initialize the Peer
	if conn == nil || conn.RemoteAddr() == nil {
		return
	}
	peer, ok := cl.atomicGetPeer(conn.RemoteAddr().String())
	if !ok {
		// This internally calls messageHandler in a separate goRoutine
		cl.SetupPeerConnections(conn.RemoteAddr().(*net.TCPAddr), conn)
		return
	}

	util.TPrintf("~~~ Got a connection! ~~~\n")
	for {
		// Process the message
		buf, err := btnet.ReadMessage(conn)
		if err != nil {
			util.WPrintf("%s\n", err)
			conn.Close()
			return
		}
		peerMessage := btnet.DecodePeerMessage(buf, len(cl.torrentMeta.PieceHashes))
		util.TPrintf("Received PeerMessage, type: %v, from: %s\n", peerMessage.Type, conn.RemoteAddr().String())

		// Really anytime we receive a message we should treat this as
		// a KeepAlive message
		peer.KeepAlive <- true

		// Massive switch case that would handle incoming messages depending on message type
		if !peerMessage.KeepAlive {
			switch peerMessage.Type {
			case btnet.Choke:
				peer.SetChoking(true)
			case btnet.Unchoke:
				peer.SetChoking(false)
			case btnet.Interested:
				peer.SetInterested(true)
			case btnet.NotInterested:
				peer.SetInterested(false)
			case btnet.Have:
				peer.SetBitfieldElement(peerMessage.Index, true)
			case btnet.Bitfield:
				peer.SetBitfield(peerMessage.Bitfield)
			case btnet.Request:
				util.TPrintf("%s: received request msg\n", cl.port)
				cl.sendBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peer)
			case btnet.Piece:
				util.TPrintf("%s: received piece msg\n", cl.port)
				cl.saveBlock(int(peerMessage.Index), peerMessage.Begin, peerMessage.Length, peerMessage.Block)
			case btnet.Cancel:
				// TODO make a cancel queue and dont send out pieces if you recieve one of these
				fallthrough
			default:
				// Unsupported message
				util.WPrintf("%s: unsupported message\n", cl.port)
				// Drop the connection
				conn.Close()
			}
		}

		// Update okay to make sure that we still have a connection
		ok = conn.RemoteAddr() != nil
		if !ok {
			return
		}
	}
	util.TPrintf("%s: exiting messageHandler\n", cl.port)
}
