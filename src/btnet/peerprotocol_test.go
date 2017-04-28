package btnet

import "testing"
import "fmt"

func TestProcessMessage(t *testing.T) {
  // data := []byte{0x00, 0x00, 0x00, 0x01, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
  // ProcessMessage(data);
}

func TestKeepAliveMessage(t *testing.T) {
  // Standard keep alive (len=0)
  data := []byte{0x00, 0x00, 0x00, 0x00}
  actual := ProcessMessage(data)
  expected := PeerMessage{KeepAlive: true, Length: 0}
  if actual.Type != expected.Type ||
     actual.Length != expected.Length {
     t.Fail()
   }
}


func TestChokeMessage(t *testing.T) {
  // Standard choke message
  data := []byte{0x00, 0x00, 0x00, 0x01, 0x00}
  actual := ProcessMessage(data)
  expected := PeerMessage{Type: Choke, Length: 1}
  if actual.Type != expected.Type ||
     actual.Length != expected.Length {
    t.Fail()
  }
}

func TestUnchokeMessage(t *testing.T) {
  // Standard unchoke message
  data := []byte{0x00, 0x00, 0x00, 0x01, 0x01}
  actual := ProcessMessage(data)
  expected := PeerMessage{Type: Unchoke, Length: 1}
  if actual.Type != expected.Type ||
     actual.Length != expected.Length {
      t.Fail()
  }
}

func TestInterestedMessage(t *testing.T) {
  // Standard interested message
  data := []byte{0x00, 0x00, 0x00, 0x01, 0x02}
  actual := ProcessMessage(data)
  expected := PeerMessage{Type: Interested, Length: 1}
  if actual.Type != expected.Type ||
     actual.Length != expected.Length {
    t.Fail()
  }
}

func TestNotInterestedMessage(t *testing.T) {
  // Standard notInterested message
  data := []byte{0x00, 0x00, 0x00, 0x01, 0x03}
  actual := ProcessMessage(data)
  expected := PeerMessage{Type: NotInterested, Length: 1}
  if actual.Type != expected.Type ||
     actual.Length != expected.Length {
    t.Fail()
  }
}

func TestHaveMessage(t *testing.T) {
  // Standard notInterested message
  data := []byte{0x00, 0x00, 0x00, 0x05, 0x04, 0x00, 0x00, 0x80, 0x00}
  actual := ProcessMessage(data)
  expected := PeerMessage{Type: Have, Length: 5, Index: 32768}
  if actual.Type != expected.Type ||
     actual.Length != expected.Length ||
     actual.Index != expected.Index {
    t.Fail()
  }
}
