package btnet

import (
	"bytes"
	"encoding/binary"
	"net"
	"util"
)

const BT_PROTOCOL string = "BitTorrent protocol"

type Handshake struct {
  Pstr string
  Reserved [8]byte
  InfoHash []byte
  PeerId []byte
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
	Status   PeerStatus
	Bitfield []bool
	Addr     net.TCPAddr
  Conn     net.TCPConn
  KeepAlive chan bool
}

// Make sure to start a go routine to kill this connection
func InitializePeer(addr *net.TCPAddr, infoHash string, peerId string, bitfieldLength int, conn *net.TCPConn) Peer {
	// tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	peer := Peer{}
	// if err != nil {
	//   fmt.Println(err)
	//   return peer
	// }
	// peer.Addr = *tcpAddr
	peer.Addr = *addr
	peer.Bitfield = make([]bool, bitfieldLength)
	peer.Status.AmChoking = true
	peer.Status.AmInterested = false
	peer.Status.PeerChoking = true
	peer.Status.PeerInterested = false
  peer.KeepAlive = make(chan bool, 100)
  // Create handshake
  // Handshake{Pstr: BT_PROTOCOL, InfoHash: []byte(infoHash), PeerId: []byte(peerId)}
  data := EncodeHandshake(Handshake{Pstr: BT_PROTOCOL, InfoHash: []byte(infoHash), PeerId: []byte(peerId)})
  if (conn != nil) {
    peer.Conn = *conn
  } else {
    peer.Conn = *DoDial(addr, data)
  }

	return peer
}

func DecodeHandshake(data []byte) Handshake {

  return Handshake{}
}

func EncodeHandshake(handshake Handshake) []byte {
  pstrlen := int8(len(handshake.Pstr))
  buf := make([]byte, 49 + int(pstrlen))
  buf[0] = byte(pstrlen)

  pstr := []byte(handshake.Pstr)
  for i := 1; i < int(pstrlen) + 1; i++ {
    buf[i] = pstr[i - 1]
  }
  // skip 8 bytes
  infoHashIndex := 9 + int(pstrlen)
  // infoHash = []byte(handshake.InfoHash)
  for i:= infoHashIndex; i < infoHashIndex + 20; i++ {
    buf[i] = handshake.InfoHash[i - infoHashIndex]
  }
  peerIdIndex := infoHashIndex + 20
  for i := peerIdIndex; i < peerIdIndex + 20; i++ {
    buf[i] = handshake.PeerId[i - peerIdIndex]
  }
  return buf
}

// fill in a PeerMessage struct from an array of bytes
func DecodePeerMessage(data []byte) PeerMessage {

	// messageType := data[0]
	var length int32
	var messageType int8
	peerMessage := PeerMessage{}
	// b := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
	buf := bytes.NewReader(data)

	// First grab the length of the message sent
	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		util.EPrintf("peerprotocol DecodePeerMessage: %s\n", err)
	}
	// peerMessage.Length = int(length) // peerMessage.Length != length
	if length < 1 {
		// This is a keepalive message
		peerMessage.KeepAlive = true
		return peerMessage
	}

	// Now read the message type
	err = binary.Read(buf, binary.BigEndian, &messageType)
	if err != nil {
		util.EPrintf("peerprotocol DecodePeerMessage: %s\n", err)
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
		peerMessage.Index = index
		return peerMessage
	case Bitfield:
		util.Printf("Bitfield message\n")
	case Request:
		util.Printf("Request message\n")
	case Piece:
		util.Printf("Piece message\n")
	case Cancel:
		util.Printf("Cancel message\n")
	default:
		util.Printf("Unsupported message\n")
	}

	return PeerMessage{}
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
		// if err != nil {
		//   fmt.Println("binary.Write failed:", err)
		// }
		// err = binary.Write(buf, binary.BigEndian, msg.Type)
		// if err != nil {
		//   fmt.Println("binary.Write failed:", err)
		// }
		// fmt.Println("Encoding Choke message")
		fallthrough
	case Unchoke:
		// err = binary.Write(buf, binary.BigEndian, int32(1))
		// if err != nil {
		//   fmt.Println("binary.Write failed:", err)
		// }
		// err = binary.Write(buf, binary.BigEndian, msg.Type)
		// if err != nil {
		//   fmt.Println("binary.Write failed:", err)
		// }
		// fmt.Println("Encoding Unchoke message")
		fallthrough
	case Interested:
		// err = binary.Write(buf, binary.BigEndian, int32(1))
		// if err != nil {
		//   fmt.Println("binary.Write failed:", err)
		// }
		// err = binary.Write(buf, binary.BigEndian, msg.Type)
		// if err != nil {
		//   fmt.Println("binary.Write failed:", err)
		// }
		// fmt.Println("Encoding Interested message")
		fallthrough
	case NotInterested:
		err = binary.Write(buf, binary.BigEndian, int32(1))
		if err != nil {
			util.EPrintf("peerprotocol EncodePeerMessage: %s\n", err)
		}
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		if err != nil {
			util.EPrintf("peerprotocol EncodePeerMessage: %s\n", err)
		}
		// fmt.Println("Encoding NotIntested message")
	case Have:
		err = binary.Write(buf, binary.BigEndian, int32(5))
		if err != nil {
			util.EPrintf("peerprotocol EncodePeerMessage: %s\n", err)
		}
		err = binary.Write(buf, binary.BigEndian, msg.Type)
		if err != nil {
			util.EPrintf("peerprotocol EncodePeerMessage: %s\n", err)
		}
		err = binary.Write(buf, binary.BigEndian, msg.Index)
		if err != nil {
			util.EPrintf("peerprotocol EncodePeerMessage: %s\n", err)
		}
		util.Printf("Encoding Have message")
	case Bitfield:
		util.Printf("Encoding BitField message")
	case Request:
		util.Printf("Encoding Request message")
	case Piece:
		util.Printf("Encoding Piece message")
	case Cancel:
		util.Printf("Encoding Cancel message")
	default:
		util.Printf("Something went wrong")
		return []byte{}
	}
	// fmt.Printf("% x", buf.Bytes())

	return buf.Bytes()
}
