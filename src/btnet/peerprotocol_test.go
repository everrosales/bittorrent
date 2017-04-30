package btnet

import "testing"
// import "fmt"
import "util"

// Byte encodings of messages
var KeepAliveBytes []byte = []byte{0x00, 0x00, 0x00, 0x00}
var ChokeBytes []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x00}
var UnchokeBytes []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x01}
var InterestedBytes []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x02}
var NotInterestedBytes []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x03}
var HaveBytes []byte = []byte{0x00, 0x00, 0x00, 0x05, 0x04, 0x00, 0x00, 0x80, 0x00}

// PeerMessage structs of messages
var KeepAliveMsg PeerMessage = PeerMessage{KeepAlive: true}
var ChokeMsg PeerMessage = PeerMessage{Type: Choke}
var UnchokeMsg PeerMessage = PeerMessage{Type: Unchoke}
var InterestedMsg PeerMessage = PeerMessage{Type: Interested}
var NotInterestedMsg PeerMessage = PeerMessage{Type: NotInterested}
var HaveMsg PeerMessage = PeerMessage{Type: Have, Index: 32768}


func TestProcessMessage(t *testing.T) {
  // data := []byte{0x00, 0x00, 0x00, 0x01, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
  // ProcessMessage(data);
}

func TestDecodeKeepAliveMessage(t *testing.T) {
  // Standard keep alive (len=0)
  actual := DecodePeerMessage(KeepAliveBytes)
  expected := KeepAliveMsg
  if actual.KeepAlive != expected.KeepAlive {
     t.Fail()
   }
}


func TestDecodeChokeMessage(t *testing.T) {
  // Standard choke message
  actual := DecodePeerMessage(ChokeBytes)
  expected := ChokeMsg
  if actual.Type != expected.Type {
    t.Fail()
  }
}

func TestDecodeUnchokeMessage(t *testing.T) {
  // Standard unchoke message
  actual := DecodePeerMessage(UnchokeBytes)
  expected := UnchokeMsg
  if actual.Type != expected.Type {
      t.Fail()
  }
}

func TestDecodeInterestedMessage(t *testing.T) {
  // Standard interested message
  actual := DecodePeerMessage(InterestedBytes)
  expected := InterestedMsg
  if actual.Type != expected.Type {
    t.Fail()
  }
}

func TestDecodeNotInterestedMessage(t *testing.T) {
  // Standard notInterested message
  actual := DecodePeerMessage(NotInterestedBytes)
  expected := NotInterestedMsg
  if actual.Type != expected.Type {
    t.Fail()
  }
}

func TestDecodeHaveMessage(t *testing.T) {
  // Standard notInterested message
  actual := DecodePeerMessage(HaveBytes)
  expected := HaveMsg
  if actual.Type != expected.Type ||
     actual.Index != expected.Index {
    t.Fail()
  }
}

func TestEncodeKeepAliveMessage(t *testing.T) {
  actual := EncodePeerMessage(KeepAliveMsg)
  expected := KeepAliveBytes
  if !util.ByteArrayEquals(actual, expected) {
    t.Fail()
  }
}

func TestEncodeChokeMessage(t *testing.T) {
  actual := EncodePeerMessage(ChokeMsg)
  expected := ChokeBytes
  if !util.ByteArrayEquals(actual, expected) {
    t.Fail()
  }
}
