package btnet

import (
	"testing"
	"util"
)

var HandshakeMsg Handshake = Handshake{
  Pstr: "BitTorrent protocol",
  InfoHash: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
                   0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19},
  PeerId: []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
                 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}}


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

func init() {
	util.Debug = util.None
}

func TestProcessMessage(t *testing.T) {
	// data := []byte{0x00, 0x00, 0x00, 0x01, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
	// ProcessMessage(data);
}

func TestEncodeDecodeHandshake(t *testing.T) {
  util.StartTest("Testing EncodeDecodeHandshake...")
  handshake := HandshakeMsg
  data := EncodeHandshake(handshake)
  decodedHandshake := DecodeHandshake(data)
  if (!util.ByteArrayEquals(handshake.InfoHash, decodedHandshake.InfoHash) ||
      !util.ByteArrayEquals(handshake.PeerId, decodedHandshake.PeerId) ||
      handshake.Pstr != decodedHandshake.Pstr) {
    t.Fail()
  }
  util.EndTest()
}

func TestDecodeHandshake(t *testing.T) {

}

func TestDecodeKeepAliveMessage(t *testing.T) {
	util.StartTest("Testing KeepAlive message...")
	// Standard keep alive (len=0)
	actual := DecodePeerMessage(KeepAliveBytes)
	expected := KeepAliveMsg
	if actual.KeepAlive != expected.KeepAlive {
		t.Fatalf("Expected KeepAlive != actual KeepAlive")
	}
	util.EndTest()
}

func TestDecodeChokeMessage(t *testing.T) {
	util.StartTest("Testing Choke message...")
	// Standard choke message
	actual := DecodePeerMessage(ChokeBytes)
	expected := ChokeMsg
	if actual.Type != expected.Type {
		t.Fatalf("Expected type != actual type")
	}
	util.EndTest()
}

func TestDecodeUnchokeMessage(t *testing.T) {
	util.StartTest("Testing Unchoke message...")
	// Standard unchoke message
	actual := DecodePeerMessage(UnchokeBytes)
	expected := UnchokeMsg
	if actual.Type != expected.Type {
		t.Fatalf("Expected type != actual type")
	}
	util.EndTest()
}

func TestDecodeInterestedMessage(t *testing.T) {
	util.StartTest("Testing Interested message...")
	// Standard interested message
	actual := DecodePeerMessage(InterestedBytes)
	expected := InterestedMsg
	if actual.Type != expected.Type {
		t.Fatalf("Expected type != actual type")
	}
	util.EndTest()
}

func TestDecodeNotInterestedMessage(t *testing.T) {
	util.StartTest("Testing NotInterested message...")
	// Standard notInterested message
	actual := DecodePeerMessage(NotInterestedBytes)
	expected := NotInterestedMsg
	if actual.Type != expected.Type {
		t.Fatalf("Expected type != actual type")
	}
	util.EndTest()
}

func TestDecodeHaveMessage(t *testing.T) {
	util.StartTest("Testing Have message...")
	// Standard notInterested message
	actual := DecodePeerMessage(HaveBytes)
	expected := HaveMsg
	if actual.Type != expected.Type || actual.Index != expected.Index {
		t.Fatalf("Expected != actual")
	}
	util.EndTest()
}

func TestEncodeKeepAliveMessage(t *testing.T) {
	util.StartTest("Testing encoding KeepAlive message...")
	actual := EncodePeerMessage(KeepAliveMsg)
	expected := KeepAliveBytes
	if !util.ByteArrayEquals(actual, expected) {
		t.Fatalf("Expected != actual")
	}
	util.EndTest()
}

func TestEncodeChokeMessage(t *testing.T) {
	util.StartTest("Testing encoding Choke message...")
	actual := EncodePeerMessage(ChokeMsg)
	expected := ChokeBytes
	if !util.ByteArrayEquals(actual, expected) {
		t.Fatalf("Expected != actual")
	}
	util.EndTest()
}
