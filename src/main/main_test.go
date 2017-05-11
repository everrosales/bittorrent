package main

import (
	"bytes"
	"client"
	"encoding/gob"
	"fs"
	"os"
	"testing"
	"tracker"
	"util"
)

const TestTorrentSmall = "torrent/puppy.torrent"
const SeedFileSmall = "seed/puppy.jpg"

func init() {
	util.Debug = util.None
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

func TestTwoClients(t *testing.T) {
	util.StartTest("Testing integration with one seeder and one downloader...")
	file := TestTorrentSmall
	seed := SeedFileSmall

	seederPersister := btclient.MakePersister("test1")
	downloaderPersister := btclient.MakePersister("test2")

	tr := bttracker.StartBTTracker(file, 8000)
	seeder := btclient.StartBTClient("localhost", 6666, file, seed, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", 6667, file, "", "puppy_download.jpg", downloaderPersister)

	util.Wait(2000)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	res := loadDataFromPersister(downloaderPersister)
	metadata := fs.Read(file)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	if len(res.Pieces) != len(metadata.PieceHashes) {
		util.EPrintf("Client has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
		t.Fail()
		return
	}

	util.TPrintf("piece bitmap %v\n", res.PieceBitmap)
	for i, hash := range metadata.PieceHashes {
		if res.Pieces[i].Hash() != hash {
			util.EPrintf("Piece %d did not hash correctly\n%s != %s\n", i, res.Pieces[i].Hash(), hash)
			t.Fail()
			return
		}
	}
	util.EndTest()
}

func TestThreeClients(t *testing.T) {
	util.StartTest("Testing integration with one seeder and one downloader...")
	file := "puppy.jpg.torrent"
	seed := "puppy.jpg"

	seederPersister := btclient.MakePersister("test1")
	downloaderPersister := btclient.MakePersister("test2")
	downloaderPersister2 := btclient.MakePersister("test3")

	tr := bttracker.StartBTTracker(file, 8000)
	seeder := btclient.StartBTClient("localhost", 6668, file, seed, seederPersister)
	downloader := btclient.StartBTClient("localhost", 6669, file, "", downloaderPersister)
	downloader2 := btclient.StartBTClient("localhost", 6670, file, "", downloaderPersister2)

	util.Wait(2000)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()
	downloader2.Kill()

	res := loadDataFromPersister(downloaderPersister)
	metadata := fs.Read(file)

	res2 := loadDataFromPersister(downloaderPersister2)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)
	os.Remove(downloaderPersister2.Path)

	// t.Fail()

	if len(res.Pieces) != len(metadata.PieceHashes) {
		util.EPrintf("Client1: has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
		t.Fail()
		return
	}

	if len(res2.Pieces) != len(metadata.PieceHashes) {
		util.EPrintf("Client2: has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
		t.Fail()
		return
	}

	util.TPrintf("piece bitmap %v\n", res.PieceBitmap)
	for i, hash := range metadata.PieceHashes {
		if res.Pieces[i].Hash() != hash {
			util.EPrintf("Client1: Piece %d did not hash correctly\n%s != %s\n", i, res.Pieces[i].Hash(), hash)
			t.Fail()
			return
		}
	}

	util.TPrintf("piece bitmap %v\n", res2.PieceBitmap)
	for i, hash := range metadata.PieceHashes {
		if res2.Pieces[i].Hash() != hash {
			util.EPrintf("Client2: Piece %d did not hash correctly\n%s != %s\n", i, res.Pieces[i].Hash(), hash)
			t.Fail()
			return
		}
	}
	util.EndTest()
}
