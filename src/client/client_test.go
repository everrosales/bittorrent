package btclient

import (
	"btnet"
	"net"
	"testing"
	"time"
	"util"
)

func init() {
	util.Debug = util.Trace
}

// Helpers
func makeTestClient(port int) *BTClient {
	persister := MakePersister("/tmp/persister/tclient.p")
	return StartBTClient("localhost", port, "../main/test.torrent", "", persister)
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
	util.Printf("WARNING Expecting: \n\t[ERROR] labtcp ReadHandshake: EOF\n\t[ERROR] Badly formatted data\n")
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

	util.Wait(1000)
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
	connection = btnet.DoDial(tcpAddr, data)
	util.Wait(100)
	_, ok := cl.peers[connection.LocalAddr().String()]
	if !ok {
		util.EPrintf("A peer should be connected\n")
		t.Fail()
		return
	}

	msg := btnet.PeerMessage{Type: btnet.Interested}
	data = btnet.EncodePeerMessage(msg)
	// util.Printf("Making a connection\n")
	util.Wait(100)
	_, ok = cl.peers[connection.LocalAddr().String()]
	if !ok {
		util.EPrintf("Client should be connected\n")
		t.Fail()
		return
	}

	util.Wait(4000)
	if len(cl.peers) > 0 {
		util.EPrintf("There should be no peers connected\n")
		t.Fail()
	} else {
		// util.Printf("Missing peer: %v\n", connection.LocalAddr())
		// t.Fail()
		util.EndTest()
	}
}

func TestTwoPeers(t *testing.T) {
	util.StartTest("Testing two peers...")
	first := makeTestClient(6668)
	util.Wait(1000)
	util.Printf("Started peer 1\n")
	second := makeTestClient(6669)
	util.Printf("Started peer 2\n")
	util.Wait(1000)
    tcpAddr, _ := net.ResolveTCPAddr("tcp", "localhost:6669")
    first.SendPeerMessage(tcpAddr, btnet.PeerMessage{KeepAlive: true})
    util.Wait(5000)
	// TODO: Make sure they do something interest
    if len(second.peers) < 1 {
        t.Fatalf("second peer list does not include the first")
    }

	first.Kill()
	second.Kill()
	util.EndTest()
}
