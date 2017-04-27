package btnet
// labtcp

import "net"
import "fmt"
import "bufio"
import "time"

func StartTCPServer(addr string, handler func(net.Conn)) {
  ln, err := net.Listen("tcp", addr)
  if err != nil {
    // complain about things dying
    fmt.Println(err)
    // fmt.Println("Something went wrong")
  }
  go func(ln net.Listener) {
    conn, err := ln.Accept()
    if err != nil {
      // complain about a thing
    }
    go handler(conn)
  }(ln)
}

func SendHelloMsg(host string, port string) {
  conn, err := net.DialTimeout("tcp", "golang.org:80", time.Millisecond * 500)
  if err != nil {
  	// handle error
  }
  fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
  // status, err := bufio.NewReader(conn).ReadString('\n')
}

func doDial(addr *net.TCPAddr, data []byte) string {
  conn, err := net.DialTCP("tcp", nil, addr)
  if err != nil {
    // Cry
  }
  // fmt.Fprintf(conn, data...)
  conn.Write(data)
  response, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    // Cry
  }
  return response

}
