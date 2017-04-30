package bttracker

// tracker test
// TODO: add tests

import (
	"testing"
	"time"
)

func makeTestTracker() *BTTracker {
	return StartBTTracker("../main/test.torrent", 8000)
}

func TestMakeTracker(t *testing.T) {
	tr := makeTestTracker()
	<-time.After(time.Millisecond * 10000)
	tr.Kill()
}
