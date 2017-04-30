package btnet

import "testing"
import "fmt"
import "net"
// import "strconv"
import "util"
// import "strings"


func testTCPHandler(conn net.Conn) {
  // Assume this is a TCP connection
  tcpConn := conn.(*net.TCPConn)
  b := make([]byte, 128)
  _, err := tcpConn.Read(b)
  if err != nil {
    // do something
    fmt.Println(err)
  }
  // fmt.Println(string(b) + "bytesRead: " + strconv.Itoa(bytesRead))
  data := []byte{0x00, 0x00, 0x00, 0x05, 0x04, 0x00, 0x00, 0x80, 0x00}
  tcpConn.Write([]byte(data))
  tcpConn.Close()
}

func TestTCP(t *testing.T) {
  fmt.Println("-----------------\nRunning: TestTCP")
  StartTCPServer("localhost:6666", testTCPHandler)
  servAddr := "localhost:6666"
  tcpAddr, _ := net.ResolveTCPAddr("tcp", servAddr)
  // Send an interested msg
  data := []byte{0x00, 0x00, 0x00, 0x01, 0x02}

  conn := DoDial(tcpAddr,data)
  fmt.Println(data)
  actual := ReadMessage(conn)
  expected :=  []byte{0x00, 0x00, 0x00, 0x05, 0x04, 0x00, 0x00, 0x80, 0x00}

  if !util.ByteArrayEquals(expected, actual) {
    t.Fail()
    fmt.Printf("Expected: ", expected ," Got: ", actual)
  }
  fmt.Println("Passed\n-----------------")
}

// TODO: theres a lot more logic in DoDial now... we should really test it
