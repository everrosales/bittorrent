package btclient

import "testing"
import "time"

func makeTestClient() *BTClient {
  persister := MakePersister()
  return StartBTClient("127.0.0.1", "8000", persister)
}

func TestMakeClient(t *testing.T) {
  client := makeTestClient()
  <- time.After(time.Millisecond * 10000)
  client.Kill()
}
