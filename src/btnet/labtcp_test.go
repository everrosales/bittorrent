package btnet

import (
	"net"
	"testing"
	"util"
)

func init() {
	util.Debug = util.None
}

// Test
func testTCPHandler(conn net.Conn) {
	// Assume this is a TCP connection
	tcpConn := conn.(*net.TCPConn)
	b := make([]byte, 128)
	_, err := tcpConn.Read(b)
	if err != nil {
		// do something
		util.EPrintf("Error reading: %s\n", err)
	}
	// fmt.Println(string(b) + "bytesRead: " + strconv.Itoa(bytesRead))
	data := []byte{0x00, 0x00, 0x00, 0x05, 0x04, 0x00, 0x00, 0x80, 0x00}
	tcpConn.Write([]byte(data))
	tcpConn.Close()
}

func TestTCP(t *testing.T) {
	util.StartTest("Test TCP...")
	StartTCPServer("localhost:6666", testTCPHandler)
	servAddr := "localhost:6666"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", servAddr)
	// Send an interested msg
	data := []byte{0x00, 0x00, 0x00, 0x01, 0x02}

	conn := DoDial(tcpAddr, data)
	util.TPrintf("data: %v\n", data)
	actual := ReadMessage(conn)
	expected := []byte{0x00, 0x00, 0x00, 0x05, 0x04, 0x00, 0x00, 0x80, 0x00}

	if !util.ByteArrayEquals(expected, actual) {
		t.Fatalf("Expected %s, got %s\n", expected, actual)
	}
	util.EndTest()
}


func TestSendPeerMessage(t *testing.T) {
  util.StartTest("Test SendPeerMessage...")
  sendPeerMessageHandler := func(conn net.Conn) {
    tcpConn := conn.(*net.TCPConn)
    b := ReadMessage(tcpConn)
    util.Printf("Message: %v\n", b)
  }

  servAddr := "localhost:6667"
  StartTCPServer(servAddr, sendPeerMessageHandler)
  // msg := PeerMessage{KeepAlive: true}
  // addr, _ := net.ResolveTCPAddr("tcp", servAddr)
  // addr := tcpAddr.(*net.Addr)
  // SendPeerMessage(addr, msg)
  util.Wait(500)
  // msg = PeerMessage{}
  util.EndTest()
}


// TODO: theres a lot more logic in DoDial now... we should really test it
