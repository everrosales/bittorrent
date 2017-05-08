package btnet

import (
	"reflect"
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
var BitfieldBytes []byte = []byte{0x00, 0x00, 0x00, 0x03, 0x05, 0x67, 0x8f}

//																		msg Length				 type				index
var RequestBytes []byte = []byte{0x00, 0x00, 0x00, 0x0d, 0x06, 0x00, 0x00, 0x00, 0x0b,
	0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08}

// 													  	   msg length					 type       index
var PieceBytes []byte = []byte{0x00, 0x00, 0x00, 0x11, 0x07, 0x00, 0x00, 0x00, 0x0b,
	0x00, 0x00, 0x01, 0x00,
	0xde, 0xad, 0xbe, 0xef, 0x00, 0x01, 0x02, 0x04}

//																		msg Length			   type				index
var CancelBytes []byte = []byte{0x00, 0x00, 0x00, 0x0d, 0x08, 0x00, 0x00, 0x00, 0x0b,
	0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08}

// PeerMessage structs of messages
var KeepAliveMsg PeerMessage = PeerMessage{KeepAlive: true}
var ChokeMsg PeerMessage = PeerMessage{Type: Choke}
var UnchokeMsg PeerMessage = PeerMessage{Type: Unchoke}
var InterestedMsg PeerMessage = PeerMessage{Type: Interested}
var NotInterestedMsg PeerMessage = PeerMessage{Type: NotInterested}
var HaveMsg PeerMessage = PeerMessage{Type: Have, Index: 32768}
var BitfieldMsg PeerMessage = PeerMessage{Type: Bitfield,
	Bitfield: []bool{false, true, true, false, false, true, true, true,
		true, false, false, false, true, true, true, true}}
var RequestMsg PeerMessage = PeerMessage{Type: Request, Index: 11, Begin: 256, Length: 264}
var PieceMsg PeerMessage = PeerMessage{Type: Piece, Index: 11, Begin: 256,
	Block: []byte{0xde, 0xad, 0xbe, 0xef, 0x00, 0x01, 0x02, 0x04}}
var CancelMsg PeerMessage = PeerMessage{Type: Cancel, Index: 11, Begin: 256, Length: 264}

func init() {
	util.Debug = util.None
}

func TestKeepAliveMessage(t *testing.T) {
	runMessageTests("KeepAlive", KeepAliveMsg, KeepAliveBytes, t)
}

func TestChokeMessage(t *testing.T) {
	runMessageTests("Choke", ChokeMsg, ChokeBytes, t)
}

func TestEncodeUnchokeMessage(t *testing.T) {
	runMessageTests("Unchoke", UnchokeMsg, UnchokeBytes, t)
}

func TestInterestedMessage(t *testing.T) {
	runMessageTests("Interested", InterestedMsg, InterestedBytes, t)
}

func TestNotInterestedMessage(t *testing.T) {
	runMessageTests("NotInterested", NotInterestedMsg, NotInterestedBytes, t)
}

func TestHaveMessage(t *testing.T) {
	runMessageTests("Have", HaveMsg, HaveBytes, t)
}

func TestBitfield(t *testing.T) {
	runMessageTests("Bitfield", BitfieldMsg, BitfieldBytes, t)
}

func TestRequestMessage(t *testing.T) {
	runMessageTests("Request", RequestMsg, RequestBytes, t)
}

func TestPieceMessage(t *testing.T) {
	runMessageTests("Piece", PieceMsg, PieceBytes, t)
}

func TestCancelMessage(t *testing.T) {
	runMessageTests("Cancel", CancelMsg, CancelBytes, t)
}

func runMessageTests(testname string, msg PeerMessage, bytes []byte, t *testing.T) {
	runEncodeTest(testname, msg, bytes, t)
	runDecodeTest(testname, bytes, msg, t)
}

func runDecodeTest(testname string, input []byte, expected PeerMessage, t *testing.T) {
	util.StartTest("Testing decode " + testname + " message...")
	actual := DecodePeerMessage(input)
	if !reflect.DeepEqual(actual, expected) {
		util.Printf("%v, %v", actual, expected)
		t.Fatalf("expected != actual")
	}
	util.EndTest()
}

func runEncodeTest(testname string, input PeerMessage, expected []byte, t *testing.T) {
	util.StartTest("Testing encode " + testname + " message...")
	actual := EncodePeerMessage(input)
	if !util.ByteArrayEquals(actual, expected) {
		util.Printf("%v, %v", actual, expected)
		t.Fatalf("expected != actual")
	}
	util.EndTest()
}

// TODO: Peer Protocol now handles initializing peers. We should write
//			 a few tests for that.
