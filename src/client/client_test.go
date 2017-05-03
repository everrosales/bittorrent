package btclient

import "testing"
import "time"
import "util"
import "btnet"
import "net"
import "fmt"

func makeTestClient(port string) *BTClient {
  persister := MakePersister()
  return StartBTClient("localhost", "6666", "../main/test.torrent", persister)
}

func TestMakeClient(t *testing.T) {
  util.TestStartPrintf("TestMakeClient")
  client := makeTestClient("6666")
  <- time.After(time.Millisecond * 1000)
  client.Kill()
  util.TestFinishPrintf("Passed")
}

func TestClientTCPServer(t *testing.T) {
  util.TestStartPrintf("TestingClientTCPServer")
  client := makeTestClient("6667")
  // TODO: We should have a ready signal that we can check to see if
  //       the client is ready to start
  // Give the client some time to start up
  <- time.After(time.Millisecond * 500)

  // Test one of each messageType
  servAddr := "localhost:6667"
  tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
  if err != nil {
    fmt.Println(err)
    t.Fail()
  }

  // Sending KeepAlive
  msg := btnet.PeerMessage{Type: btnet.Interested}
  data := btnet.EncodePeerMessage(msg)
  fmt.Println("encoded data: %v", data)
  connection := btnet.DoDial(tcpAddr, data)
  status, ok := client.peers[connection.LocalAddr()]
  if !ok {
    fmt.Println("Its not there")
  }

  fmt.Println(status)
  client.Kill()

  util.TestFinishPrintf("Passed")
}

func TestTwoPeers(t *testing.T) {
  util.TestStartPrintf("TestTwoPeers")
  first := makeTestClient("6668")
  second := makeTestClient("6669")

  // TODO: Make sure they do something interest

  first.Kill()
  second.Kill()
  util.TestFinishPrintf("Passed")
}
