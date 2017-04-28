package btnet

import "net"
import "fmt"

type MessageType int

const (
  Choke MessageType = iota  // 0
  Unchoke // 1
  Interested // 2
  NotInterested // 3
  Have // 4
  Bitfield // 5
  Request // 6
  Piece // 7
  Cancel // 8
)

type PeerMessage struct {
  Type MessageType
  Index int
  Begin int
  Length int
  Bitfield []bool
  Block []byte
  // Zero length messages are keep alive messages and have no type
  KeepAlive bool
}

type PeerStatus struct {
  AmChoking bool      // This client is choking this peer
  AmInterested bool   // This client is interested in this peer
  PeerChoking bool    // This peer is choking this client
  PeerInterested bool   // This peer is interested in this client
}

type Peer struct {
  Status PeerStatus
  Bitfield []bool
  Addr  net.TCPAddr
}

// addr of format "192.168.1.0:8080"
func InitializePeer(addr string, bitfieldLength int) Peer {
  tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
  peer := Peer{}
  if err != nil {
    fmt.Println(err)
    return peer
  }
  peer.Addr = *tcpAddr
  peer.Bitfield = make([]bool, bitfieldLength)
  peer.Status.AmChoking = true
  peer.Status.AmInterested = false
  peer.Status.PeerChoking = true
  peer.Status.PeerInterested = false

  return peer
}

// fill in a PeerMessage struct from an array of bytes
func ProcessMessage(data []byte) PeerMessage {
  if len(data) < 1 {
    // This is a keepalive message
    return PeerMessage{KeepAlive: true}
  }
  messageType := data[0]

}
