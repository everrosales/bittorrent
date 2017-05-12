package test

import (
	"bytes"
	"client"
	"encoding/gob"
	"fs"
	"os"
	"sync/atomic"
	"testing"
	"time"
	"util"
)

const (
	TorrentS     = "torrent/puppy.torrent"
	SeedS        = "seed/puppy.jpg"
	PortS        = 8000
	TorrentM     = "torrent/pupper.torrent"
	SeedM        = "seed/pupper.png"
	PortM        = 8001
	WaitForDeath = 500
)

var curPort int32

func init() {
	util.Debug = util.None
	curPort = 6666
}

func nextPort() int {
	return int(atomic.AddInt32(&curPort, 1))
}

func generateOutFile() string {
	return "out/output_" + util.GenerateRandStr(5)
}

func makePersister() *btclient.Persister {
	return btclient.MakePersister("out/persist_" + util.GenerateRandStr(5))
}

func loadDataFromPersister(ps *btclient.Persister) btclient.BTClient {
	data := ps.ReadState()
	cl := btclient.BTClient{}

	if data == nil || len(data) < 1 { // bootstrap without any state?
		return cl
	}
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&cl.Pieces)
	d.Decode(&cl.PieceBitmap)
	return cl
}

// fails if test times out
func waitUntilDone(t *testing.T, all bool, cls ...*btclient.BTClient) {
	timer := time.NewTimer(time.Second * 60)
	for {
		allDone := true
		for _, cl := range cls {
			select {
			case <-timer.C:
				t.Fatalf("Downloading timed out")
			default:
				done := cl.CheckDone()
				if !all && done {
					return
				}
				if !done {
					allDone = false
				}
			}
		}
		if allDone {
			return
		}
		util.Wait(100)
	}
}

// fails if test times out
func waitUntilStarted(t *testing.T, cls ...*btclient.BTClient) {
	timer := time.NewTimer(time.Second * 30)
	for {
		allStarted := true
		for _, cl := range cls {
			select {
			case <-timer.C:
				t.Fatalf("Downloading timed out")
			default:
				if util.AllFalse(cl.AtomicGetBitmap()) {
					allStarted = false
				}
			}
		}
		if allStarted {
			return
		}
		util.Wait(100)
	}
}

func checkDownloadResult(t *testing.T, res btclient.BTClient, file string, seed string, output string) {
	metadata := fs.Read(file)
	if len(res.Pieces) != len(metadata.PieceHashes) {
		t.Fatalf("Client has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
	}

	for i, hash := range metadata.PieceHashes {
		if res.Pieces[i].Hash() != hash {
			t.Fatalf("Piece %d did not hash correctly\n%s != %s\n", i, res.Pieces[i].Hash(), hash)
		}
	}
	same, err := util.CompareFiles(seed, output)
	if err != nil || !same {
		t.Fatalf("Seed file and downloaded file don't match: %s", err.Error())
	}

	os.Remove(output)

}
