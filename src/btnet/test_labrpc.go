package btnet

import "testing"

func TestSendHelloMsg(*testing.T) {
  testHandler := func(conn Conn) {
    fmt.Println(conn)
    fmt.Println(conn.)
  }

  btnet.StartTCPServer("8080", testHander)

}
