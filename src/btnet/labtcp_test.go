package btnet

import "testing"
import "fmt"
import "net"
// import "strconv"
import "strings"


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
  data := "I got you fam\n"
  tcpConn.Write([]byte(data))
  tcpConn.Close()
}

func TestTCP(t *testing.T) {
  StartTCPServer("localhost:6666", testTCPHandler)
  servAddr := "localhost:6666"
  tcpAddr, _ := net.ResolveTCPAddr("tcp", servAddr)
  data := "send me pls\n"
  actual := doDial(tcpAddr,[]byte(data))
  expected := "I got you fam\n"
  if strings.Compare(expected, actual) != 0 {
    t.Fail()
    fmt.Printf("Expected: " + expected + " Got: " + actual)
  }
}
