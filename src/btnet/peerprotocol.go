package btnet

import "net"

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
}

type PeerStatus struct {
  AmChoking bool      // This client is choking this peer
  AmInterested bool   // This client is interested in this peer
  PeerChoking bool    // This peer is choking this client
  PeerInterest bool   // This peer is interested in this client
}

type Peer struct {
  Status PeerStatus
  Bitfield []bool
  Addr  net.TCPAddr
}
