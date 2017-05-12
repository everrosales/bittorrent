package btnet

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"
	"util"
)

const BT_PROTOCOL string = "BitTorrent protocol"

type Handshake struct {
	Pstr     string
	Reserved [8]byte
	InfoHash []byte
	PeerId   []byte
}

type MessageType int8

const (
	Choke         MessageType = iota // 0
	Unchoke                          // 1
	Interested                       // 2
	NotInterested                    // 3
	Have                             // 4
	Bitfield                         // 5
	Request                          // 6
	Piece                            // 7
	Cancel                           // 8
)

type PeerMessage struct {
	Type     MessageType
	Index    int32
	Begin    int
	Length   int
	Bitfield []bool
	Block    []byte

	BlockLength int
	// Zero length messages are keep alive messages and have no type
	KeepAlive bool
}

type PeerStatus struct {
	AmChoking      bool // This client is choking this peer
	AmInterested   bool // This client is interested in this peer
	PeerChoking    bool // This peer is choking this client
	PeerInterested bool // This peer is interested in this client
}

type Peer struct {
	mu        sync.RWMutex
	Status    PeerStatus
	Bitfield  []bool
	Addr      net.TCPAddr
	Conn      net.TCPConn
	MsgQueue  chan PeerMessage
	KeepAlive chan bool
}

func (p *Peer) GetBitfield() []bool {
	p.mu.RLock()
	result := make([]bool, len(p.Bitfield))
	copy(result, p.Bitfield)
	defer p.mu.RUnlock()
	return result
}

func (p *Peer) GetStatus() PeerStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status
}

func (p *Peer) SetBitfield(arr []bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(arr) != len(p.Bitfield) {
		panic("Bitfield lengths don't match")
	}
	copy(p.Bitfield, arr)
}

func (p *Peer) SetBitfieldElement(index int32, val bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Bitfield[index] = val
}

func (p *Peer) SetChoking(val bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status.PeerChoking = val
}

func (p *Peer) SetInterested(val bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status.PeerInterested = val
}

// Make sure to start a go routine to kill this connection
func InitializePeer(addr *net.TCPAddr, infoHash string, peerId string, bitfieldLength int, conn *net.TCPConn, pieceBitmap []bool) *Peer {
	// tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	peer := Peer{}
	// if err != nil {
	//   fmt.Println(err)
	//   return peer
	// }
	// peer.Addr = *tcpAddr
	peer.Addr = *addr
	peer.Bitfield = make([]bool, bitfieldLength)
	peer.Status.AmChoking = false
	peer.Status.AmInterested = false
	peer.Status.PeerChoking = false
	peer.Status.PeerInterested = false
	peer.MsgQueue = make(chan PeerMessage, 100)
	peer.KeepAlive = make(chan bool, 100)
	// Create handshake
	// Handshake{Pstr: BT_PROTOCOL, InfoHash: []byte(infoHash), PeerId: []byte(peerId)}
	if conn != nil && conn.RemoteAddr() != nil {
		// This happens if we are not the ones initializing the communication
		data, err := ReadHandshake(conn)
		if err != nil {
			util.WPrintf("%s\n", err)
			conn.Close()
			return nil
		}
		handshake := DecodeHandshake(data)
		// TODO: Process the handshake and drop connection if needed
		if len(handshake.InfoHash) != 20 {
			util.TPrintf("CR: BAD handshake, closing the connection\n")
			// Badly formatted hsandshake, dont make the connection stick
			conn.Close()
			return nil
		}
		// TODO: Send bitfield message
		message := PeerMessage{
			Type:     Bitfield,
			Bitfield: pieceBitmap}
		util.TPrintf("Enqueuing bitfield message %v\n", pieceBitmap)
		peer.MsgQueue <- message
		// cl.SendPeerMessage(&peer.Addr, message)
		peer.Conn = *conn
	} else {
		handshake := Handshake{Pstr: BT_PROTOCOL, InfoHash: []byte(infoHash), PeerId: []byte(peerId)}
		data := EncodeHandshake(handshake)
		// Sending data
		util.TPrintf("Sending Handshake\n")

		conn, err := DoDial(addr, data)
		if err != nil {
			return nil
		}
		peer.Conn = *conn

		message := PeerMessage{
			Type:     Bitfield,
			Bitfield: pieceBitmap}
		util.TPrintf("Enqueuing bitfield message %v\n", pieceBitmap)
		peer.MsgQueue <- message
		// Read bitfield message that gets sent back
	}

	return &peer
}

func DecodeHandshake(data []byte) Handshake {
	if len(data) < 1 {
		util.WPrintf("Badly formatted data\n")
		return Handshake{}
	}

	pstrbuf := make([]byte, 1)
	pstrbuf[0] = data[0]
	var pstrLen uint8
	pstrLenDecodeBuf := bytes.NewReader(pstrbuf)
	errBinary := binary.Read(pstrLenDecodeBuf, binary.BigEndian, &pstrLen)
	if errBinary != nil {
		util.WPrintf("labtcp DecodeHandshake: %s\n", errBinary)
	}

	if len(data) < (49 + int(pstrLen)) {
		util.WPrintf("Badly formatted data\n")
		return Handshake{}
	}

	// Decode pstr
	pstr := string(data[1 : int(pstrLen)+1])

	// Decode infoHash
	infoHashIndex := pstrLen + 9
	infoHash := []byte(data[infoHashIndex : infoHashIndex+20])

	// Decode peerId
	peerIdIndex := infoHashIndex + 20
	peerId := []byte(data[peerIdIndex : peerIdIndex+20])

	return Handshake{Pstr: pstr, InfoHash: infoHash, PeerId: peerId}
}

func EncodeHandshake(handshake Handshake) []byte {
	pstrlen := uint8(len(handshake.Pstr))
	buf := make([]byte, 49+int(pstrlen))
	buf[0] = byte(pstrlen)

	pstr := []byte(handshake.Pstr)
	for i := 1; i < int(pstrlen)+1; i++ {
		buf[i] = pstr[i-1]
	}
	// skip 8 bytes
	infoHashIndex := 9 + int(pstrlen)
	// infoHash = []byte(handshake.InfoHash)
	for i := infoHashIndex; i < infoHashIndex+20; i++ {
		buf[i] = handshake.InfoHash[i-infoHashIndex]
	}
	peerIdIndex := infoHashIndex + 20
	for i := peerIdIndex; i < peerIdIndex+20; i++ {
		buf[i] = handshake.PeerId[i-peerIdIndex]
	}
	return buf
}

// fill in a PeerMessage struct from an array of bytes
func DecodePeerMessage(data []byte, numPieces int) PeerMessage {

	// messageType := data[0]
	var msglength int32
	var messageType int8
	peerMessage := PeerMessage{}
	// b := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
	buf := bytes.NewReader(data)

	// First grab the length of the message sent
	err := binary.Read(buf, binary.BigEndian, &msglength)
	if err != nil {
		util.WPrintf("peerprotocol DecodePeerMessage: %s\n", err)
	}
	// peerMessage.Length = int(length) // peerMessage.Length != length
	if msglength < 1 {
		// This is a keepalive message
		peerMessage.KeepAlive = true
		return peerMessage
	}

	// Now read the message type
	err = binary.Read(buf, binary.BigEndian, &messageType)
	if err != nil {
		util.WPrintf("peerprotocol DecodePeerMessage: %s\n", err)
	}
	peerMessage.Type = MessageType(messageType)

	// Now for the fun packing the of PeerMessage Struct
	switch peerMessage.Type {
	case Choke:
		// No further information needs to be parsed
		// fmt.Println("Choke message")
		// return peerMessage
		fallthrough
	case Unchoke:
		// No further information needs to be parsed
		// fmt.Println("Unchoke message")
		// return peerMessage
		fallthrough
	case Interested:
		// No further information needs to be parsed
		// fmt.Println("Interested message")
		// return peerMessage
		fallthrough
	case NotInterested:
		// No further information needs to be parsed
		// fmt.Println("NotInterested message")
		return peerMessage
	case Have:
		// fmt.Println("Have message")
		var index int32
		err = binary.Read(buf, binary.BigEndian, &index)
		checkAndPrintErr(err)
		peerMessage.Index = index
		return peerMessage
	case Bitfield:
		util.TPrintf("Decoding bitfield message\n")
		bitfield := make([]byte, (msglength - 1))
		err = binary.Read(buf, binary.BigEndian, &bitfield)
		checkAndPrintErr(err)
		peerMessage.Bitfield = util.BytesToBools(bitfield)[:numPieces]
	case Request:
		var index int32
		var begin int32
		var length int32
		err = binary.Read(buf, binary.BigEndian, &index)
		checkAndPrintErr(err)
		err = binary.Read(buf, binary.BigEndian, &begin)
		checkAndPrintErr(err)
		err = binary.Read(buf, binary.BigEndian, &length)
		checkAndPrintErr(err)
		peerMessage.Index = index
		peerMessage.Begin = int(begin)
		peerMessage.Length = int(length)
	case Piece:
		var index int32
		var begin int32
		block := make([]byte, msglength-9)
		err = binary.Read(buf, binary.BigEndian, &index)
		checkAndPrintErr(err)
		err = binary.Read(buf, binary.BigEndian, &begin)
		checkAndPrintErr(err)
		err = binary.Read(buf, binary.BigEndian, &block)
		checkAndPrintErr(err)
		peerMessage.Index = index
		peerMessage.Begin = int(begin)
		peerMessage.Block = block
	case Cancel:
		var index int32
		var begin int32
		var length int32
		err = binary.Read(buf, binary.BigEndian, &index)
		checkAndPrintErr(err)
		err = binary.Read(buf, binary.BigEndian, &begin)
		checkAndPrintErr(err)
		err = binary.Read(buf, binary.BigEndian, &length)
		checkAndPrintErr(err)
		peerMessage.Index = index
		peerMessage.Begin = int(begin)
		peerMessage.Length = int(length)
	default:
		util.WPrintf("Unsupported message\n")
		return PeerMessage{}
	}

	return peerMessage
}

func EncodePeerMessage(msg PeerMessage) []byte {
	buf := new(bytes.Buffer)

	if msg.KeepAlive {
		return []byte{0x00, 0x00, 0x00, 0x00}
	}
	// var pi float64 = math.Pi
	var err error
	// err = binary.Write(buf, binary.BigEndian, msg.Type)
	// if err != nil {
	// 	fmt.Println("binary.Write failed:", err)
	// }

	switch msg.Type {
	case Choke:
		// err = binary.Write(buf, binary.BigEndian, int32(1))
		// checkAndPrintErr(err)
		// err = binary.Write(buf, binary.BigEndian, msg.Type)
		// checkAndPrintErr(err)
		// fmt.Println("Encoding Choke message")
		fallthrough
	case Unchoke:
		// err = binary.Write(buf, binary.BigEndian, int32(1))
		// checkAndPrintErr(err)
		// err = binary.Write(buf, binary.BigEndian, msg.Type)
		// checkAndPrintErr(err)
		// fmt.Println("Encoding Unchoke message")
		fallthrough
	case Interested:
		// err = binary.Write(buf, binary.BigEndian, int32(1))
		// checkAndPrintErr(err)
		// err = binary.Write(buf, binary.BigEndian, msg.Type)
		// checkAndPrintErr(err)
		// fmt.Println("Encoding Interested message")
		fallthrough
	case NotInterested:
		err = binary.Write(buf, binary.BigEndian, int32(1))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		checkAndPrintErr(err)
		// fmt.Println("Encoding NotIntested message")
	case Have:
		err = binary.Write(buf, binary.BigEndian, int32(5))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Index)
		checkAndPrintErr(err)
	case Bitfield:
		bitFieldBuf := util.BoolsToBytes(msg.Bitfield)
		err = binary.Write(buf, binary.BigEndian, int32(len(bitFieldBuf)+1))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, bitFieldBuf)
		checkAndPrintErr(err)
	case Request:
		err = binary.Write(buf, binary.BigEndian, int32(13))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Index)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, int32(msg.Begin))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, int32(msg.Length))
	case Piece:
		err = binary.Write(buf, binary.BigEndian, int32(9+len(msg.Block)))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Index)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, int32(msg.Begin))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Block)
	case Cancel:
		err = binary.Write(buf, binary.BigEndian, int32(13))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, msg.Index)
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, int32(msg.Begin))
		checkAndPrintErr(err)
		err = binary.Write(buf, binary.BigEndian, int32(msg.Length))
	default:
		util.WPrintf("Something went wrong\n")
		return []byte{}
	}
	return buf.Bytes()
}

func checkAndPrintErr(err error) {
	if err != nil {
		util.WPrintf("peerprotocol: %s\n", err)
	}
}
