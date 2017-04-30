package btnet
// labtcp

import "net"
import "fmt"
import "bufio"
// import "time"

func StartTCPServer(addr string, handler func(net.Conn)) {
  fmt.Println("Starting the TCP Server")
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
      fmt.Println(err)
    }
    go handler(conn)
  }(ln)
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
