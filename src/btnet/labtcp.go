package btnet
// labtcp

import "net"
import "fmt"
import "bt_peer"
import "bufio"

func StartTCPServer(port string, handler func) {
  ln, err := net.Listen("tcp", port)
  if err != nil {
    // complain about things dying
  }
  go func(ln Listen) {
    conn, err := ln.Accept()
    if err != nil {
      // complain about a thing
    }
    go handler(conn)
  }(ln)
}

func SendHelloMsg(host string, port string) {
  conn, err := net.DialTimeout("tcp", "golang.org:80")
  if err != nil {
  	// handle error
  }
  fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
  status, err := bufio.NewReader(conn).ReadString('\n')
}

func doDial(addr TCPAddr, data []byte) string {
  conn, err := net.DialTCP("tcp", addr)
  if err != nil {
    // Cry
  }
  fmt.Fprintf(con, data...)
  response, err := bufio.NewReader(conn).ReadString('\n')
  return response

}
