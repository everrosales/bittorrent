package btclient

import (
	"btnet"
	"net"
	"testing"
	"time"
	"util"
)

const TestFile = "../main/torrent/test.torrent"

func init() {
	util.Debug = util.None
}

// Helpers
func makeTestClient(port int) *BTClient {
	persister := MakePersister("/tmp/persister/tclient.p")
	return StartBTClient("localhost", port, TestFile, "", "", persister)
}

// Tests
func TestMakeClient(t *testing.T) {
	util.StartTest("Testing basic starting and killing of client...")
	cl := makeTestClient(6666)
	util.Wait(1000)
	cl.Kill()
	util.EndTest()
}

func TestClientTCPServerNice(t *testing.T) {
	util.StartTest("Testing basic tcp client + nice peer")
	cl := makeTestClient(6670)
	// TODO: We should have a ready signal that we can check to see if
	//       the client is ready to start
	// Give the client some time to start up
	util.Wait(1000)

	// Test one of each messageType
	servAddr := "localhost:6670"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		cl.Kill()
		t.Fatalf("Error resolving TCP addr")
	}

	// First send handshake
	handshake := btnet.Handshake{
		Pstr: "BitTorrent protocol",
		InfoHash: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19},
		PeerId: []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}}
	data := btnet.EncodeHandshake(handshake)
	// Sending KeepAlive
	connection, err := btnet.DoDial(tcpAddr, data)
	if err != nil {
		cl.Kill()
		t.Fatalf("DoDial error: %s", err.Error())
	}
	util.Wait(100)
	_, ok := cl.getPeer(connection.LocalAddr().String())
	if !ok {
		cl.Kill()
		t.Fatalf("A peer should be connected\n")
	}

	returnedData, err := btnet.ReadMessage(connection)
	decodedMsg := btnet.DecodePeerMessage(returnedData, len(cl.torrentMeta.PieceHashes))
	if err != nil || decodedMsg.Type != 5 {
		cl.Kill()
		t.Fatalf("Did not recieve bitfield message\n%v\n", decodedMsg)
	}
	// connection.Read

	msg := btnet.PeerMessage{Type: btnet.Interested}
	data = btnet.EncodePeerMessage(msg)
	connection.Write(data)
	util.Wait(100)
	_, ok = cl.getPeer(connection.LocalAddr().String())
	if !ok {
		cl.Kill()
		t.Fatalf("Client should be connected\n")
	}
	connection.SetKeepAlive(false)

	util.Wait(6000)
	if cl.getNumPeers() > 0 {
		cl.Kill()
		t.Fatalf("There should be no peers connected\n")
	} else {
		cl.Kill()
		util.EndTest()
	}
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
		cl.Kill()
		t.Fatalf("Error resolving TCP addr")
	}

	// Send badly formatted message
	baddata := []byte{0xde, 0xad, 0xbe, 0xef}
	connection, err := btnet.DoDial(tcpAddr, baddata)
	if err != nil {
		cl.Kill()
		t.Fatalf("DoDial error: %s", err.Error())
	}
	connection.SetDeadline(time.Now().Add(500 * time.Millisecond))
	util.Wait(1000)
	if cl.getNumPeers() > 0 {
		cl.Kill()
		t.Fatalf("There should be no peers connected\n")
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
	connection, err = btnet.DoDial(tcpAddr, data)
	if err != nil {
		cl.Kill()
		t.Fatalf("DoDial error: %s", err.Error())
	}
	util.Wait(100)
	_, ok := cl.getPeer(connection.LocalAddr().String())
	if !ok {
		cl.Kill()
		t.Fatalf("A peer should be connected\n")
	}

	msg := btnet.PeerMessage{Type: btnet.Interested}
	data = btnet.EncodePeerMessage(msg)
	util.Wait(100)
	_, ok = cl.getPeer(connection.LocalAddr().String())
	if !ok {
		cl.Kill()
		t.Fatalf("Client should be connected\n")
	}

	util.Wait(5000)
	if cl.getNumPeers() > 0 {
		cl.Kill()
		t.Fatalf("There should be no peers connected\n")
	} else {
		cl.Kill()
		util.EndTest()
	}
}

func TestTwoPeers(t *testing.T) {
	util.StartTest("Testing two peers...")
	first := makeTestClient(6668)
	util.Wait(1000)
	second := makeTestClient(6669)
	util.Wait(1000)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "localhost:6669")
	first.SendPeerMessage(tcpAddr, btnet.PeerMessage{KeepAlive: true})
	util.Wait(10000)
	// TODO: Make sure they do something interest
	if second.getNumPeers() < 1 {
		first.Kill()
		second.Kill()
		t.Fatalf("second peer list does not include the first")
	}

	first.Kill()
	second.Kill()
	util.EndTest()
}
