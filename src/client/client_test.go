package btclient

import (
	"btnet"
	"net"
	"testing"
	"util"
	"time"
)

func init() {
	util.Debug = util.None
}

// Helpers
func makeTestClient(port int) *BTClient {
	persister := MakePersister("/tmp/persister/tclient.p")
	return StartBTClient("localhost", port, "../main/test.torrent", persister)
}

// Tests
func TestMakeClient(t *testing.T) {
	util.StartTest("Testing basic starting and killing of client...")
	cl := makeTestClient(6666)
	util.Wait(1000)
	cl.Kill()
	util.EndTest()
}

func TestClientTCPServer(t *testing.T) {
	util.StartTest("Testing client TCP server...")
	cl := makeTestClient(6667)
	// TODO: We should have a ready signal that we can check to see if
	//       the client is ready to start
	// Give the client some time to start up
	util.Wait(1000)

	// Test one of each messageType
	servAddr := "localhost:6667"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		t.Fatalf("Error resolving TCP addr")
	}

	// Send badly formatted message
	baddata := []byte{0xde, 0xad, 0xbe, 0xef}
	connection := btnet.DoDial(tcpAddr, baddata)
	connection.SetDeadline(time.Now().Add(500 * time.Millisecond))
	util.Wait(1000)
	if len(cl.peers) > 0 {
		util.EPrintf("There should be no peers connected\n")
		t.Fail()
	}
	connection.Close()

  util.Wait(10000)
	util.Printf("Sending second handshake\n")
	// First send handshake
	handshake := btnet.Handshake{
	  Pstr: "BitTorrent protocol",
	  InfoHash: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
	                   0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19},
	  PeerId: []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
	                 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}}
  data := btnet.EncodeHandshake(handshake)
	// Sending KeepAlive
	util.TPrintf("Encoded data: %v\n", data)
	util.Printf("Waiting for thing\n")
	connection = btnet.DoDial(tcpAddr, data)
	util.Wait(100)
	util.Printf("Done waiting for thing\n")
	// connection.Close()
	status, ok := cl.peers[connection.LocalAddr().String()]
	if ok {
		connection.Close()
		cl.Kill()
		util.TPrintf("Status: %s\n", status)
		util.EndTest()
		return
	}

	// msg := btnet.PeerMessage{Type: btnet.Interested}
	// data = btnet.EncodePeerMessage(msg)
  // // util.Printf("Making a connection\n")
	// util.Wait(1000)
	// status, ok := cl.peers[connection.LocalAddr().String()]
	// if ok {
	// 	connection.Close()
	// 	cl.Kill()
	// 	util.TPrintf("Status: %s\n", status)
	// 	util.EndTest()
	// 	return
	// }
	// util.Printf("Missing peer: %v\n", connection.LocalAddr())
	util.Printf("cl.peers: %v\n", cl.peers)
  t.Fail()
}

func TestTwoPeers(t *testing.T) {
	util.StartTest("Testing two peers...")
	first := makeTestClient(6668)
  util.Wait(1000)
  util.Printf("Started peer 1\n")
	second := makeTestClient(6669)
  util.Printf("Started peer 2\n")
  util.Wait(1000)
	// TODO: Make sure they do something interest

	first.Kill()
	second.Kill()
	util.EndTest()
}
