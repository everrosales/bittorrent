package test

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

const (
	TestTorrentSmall = "torrent/puppy.torrent"
	SeedFileSmall    = "seed/puppy.jpg"
	TestTorrentLarge = "torrent/pupper.torrent"
	SeedFileLarge    = "seed/pupper.png"
)

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

// Tests
func TestTwoClients(t *testing.T) {
	util.StartTest("Testing small file with one seeder and one downloader...")
	file := TestTorrentSmall
	seed := SeedFileSmall
	output := "out/puppy_download1.jpg"

	seederPersister := btclient.MakePersister("out/test1")
	downloaderPersister := btclient.MakePersister("out/test2")

	tr := bttracker.StartBTTracker(file, 8000)
	seeder := btclient.StartBTClient("localhost", 6666, file, seed, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", 6667, file, "", output, downloaderPersister)

	util.TPrintf("  Waiting for download to finish...\n")
	for !downloader.CheckDone() {
		util.Wait(100)
	}

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	res := loadDataFromPersister(downloaderPersister)
	metadata := fs.Read(file)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	if len(res.Pieces) != len(metadata.PieceHashes) {
		t.Fatalf("Client has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
	}

	util.TPrintf("piece bitmap %v\n", res.PieceBitmap)
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
	util.Wait(1000)
	util.EndTest()
}

func TestThreeClients(t *testing.T) {
	util.StartTest("Testing small file with one seeder and two downloaders...")
	file := TestTorrentSmall
	seed := SeedFileSmall
	output1 := "out/puppy_download2.jpg"
	output2 := "out/puppy_download3.jpg"

	seederPersister := btclient.MakePersister("out/test1")
	downloaderPersister := btclient.MakePersister("out/test2")
	downloaderPersister2 := btclient.MakePersister("out/test3")

	tr := bttracker.StartBTTracker(file, 8000)
	seeder := btclient.StartBTClient("localhost", 6668, file, seed, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", 6669, file, "", output1, downloaderPersister)
	downloader2 := btclient.StartBTClient("localhost", 6670, file, "", output2, downloaderPersister2)

	util.TPrintf("  Waiting for download to finish...\n")
	for !downloader.CheckDone() || !downloader2.CheckDone() {
		util.Wait(100)
	}

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

	if len(res.Pieces) != len(metadata.PieceHashes) {
		t.Fatalf("Client1: has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
	}

	if len(res2.Pieces) != len(metadata.PieceHashes) {
		t.Fatalf("Client2: has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
	}

	util.TPrintf("piece bitmap %v\n", res.PieceBitmap)
	for i, hash := range metadata.PieceHashes {
		if res.Pieces[i].Hash() != hash {
			t.Fatalf("Client1: Piece %d did not hash correctly\n%s != %s\n", i, res.Pieces[i].Hash(), hash)
		}
	}

	util.TPrintf("piece bitmap %v\n", res2.PieceBitmap)
	for i, hash := range metadata.PieceHashes {
		if res2.Pieces[i].Hash() != hash {
			t.Fatalf("Client2: Piece %d did not hash correctly\n%s != %s\n", i, res.Pieces[i].Hash(), hash)
		}
	}
	same, err := util.CompareFiles(seed, output1)
	if err != nil || !same {
		t.Fatalf("Client1: Seed file and downloaded file don't match: %s", err.Error())
	}
	same, err = util.CompareFiles(seed, output2)
	if err != nil || !same {
		t.Fatalf("Client2: Seed file and downloaded file don't match: %s", err.Error())
	}

	os.Remove(output1)
	os.Remove(output2)
	util.Wait(1000)
	util.EndTest()
}

func TestTwoClientsLargeFile(t *testing.T) {
	util.StartTest("Testing 36 piece file with one seeder and one downloader...")
	file := TestTorrentLarge
	seed := SeedFileLarge
	output := "out/pupper_download1.jpg"

	seederPersister := btclient.MakePersister("out/test1")
	downloaderPersister := btclient.MakePersister("out/test2")

	tr := bttracker.StartBTTracker(file, 8001)
	seeder := btclient.StartBTClient("localhost", 6671, file, seed, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", 6672, file, "", output, downloaderPersister)

	util.TPrintf("  Waiting for download to finish...\n")
	for !downloader.CheckDone() {
		util.Wait(100)
	}

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	res := loadDataFromPersister(downloaderPersister)
	metadata := fs.Read(file)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	if len(res.Pieces) != len(metadata.PieceHashes) {
		t.Fatalf("Client has %d pieces but expected %d pieces\n", len(res.Pieces), len(metadata.PieceHashes))
	}

	util.TPrintf("piece bitmap %v\n", res.PieceBitmap)
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
	util.EndTest()
}

// TODO: test with other files/piece sizes/numbers of pieces
// TODO: test with one seeder, multiple downloaders
// TODO: test with multiple seeders, multiple downloaders
// TODO: test with 1 stopped peer
// TODO: test with stopped and restarted seeder
// TODO: test with stopped and restarted downloader, test with persister
// TODO: test with stopped and restarted tracker
// TODO: test for seeding and downloading in parallel (how?)
