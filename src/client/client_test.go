package btclient

import (
	"btnet"
	"net"
	"testing"
	"util"
)

// Helpers
func makeTestClient(port int) *BTClient {
	util.Debug = util.Trace
	persister := MakePersister()
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

	// Sending KeepAlive
	msg := btnet.PeerMessage{Type: btnet.Interested}
	data := btnet.EncodePeerMessage(msg)
	util.TPrintf("Encoded data: %v\n", data)
	connection := btnet.DoDial(tcpAddr, data)
	status, ok := cl.peers[connection.LocalAddr()]
	if !ok {
		util.Printf("Missing peer")
	}

	util.TPrintf("Status: %s\n", status)
	cl.Kill()
	util.EndTest()
}

func TestTwoPeers(t *testing.T) {
	util.StartTest("Testing two peers...")
	first := makeTestClient(6668)
	second := makeTestClient(6669)

	// TODO: Make sure they do something interest

	first.Kill()
	second.Kill()
	util.EndTest()
}
