package btclient

import "testing"
import "time"

func makeTestClient() *BTClient {
  persister := MakePersister()
  return StartBTClient("localhost", "6666", persister)
}

func TestMakeClient(t *testing.T) {
  client := makeTestClient()
  <- time.After(time.Millisecond * 10000)
  client.Kill()
}
