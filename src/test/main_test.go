package test

import (
	"bytes"
	"client"
	"encoding/gob"
	"fs"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"
	"tracker"
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

// Tests
func TestOneDownloaderBasic(t *testing.T) {
	util.StartTest("Testing small file with one seeder and one downloader...")
	output := generateOutFile()
	seederPersister := makePersister()
	downloaderPersister := makePersister()

	tr := bttracker.StartBTTracker(TorrentS, PortS)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentS, SeedS, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentS, "", output, downloaderPersister)

	util.TPrintf("  Waiting for download to finish...\n")
	waitUntilDone(t, true, downloader)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	checkDownloadResult(t, res, TorrentS, SeedS, output)
	util.EndTest()
}

func TestTwoDownloaders(t *testing.T) {
	util.StartTest("Testing small file with one seeder and two downloaders...")
	output := generateOutFile()
	output2 := generateOutFile()
	seederPersister := makePersister()
	downloaderPersister := makePersister()
	downloaderPersister2 := makePersister()

	tr := bttracker.StartBTTracker(TorrentS, PortS)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentS, SeedS, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentS, "", output, downloaderPersister)
	downloader2 := btclient.StartBTClient("localhost", nextPort(), TorrentS, "", output2, downloaderPersister2)

	waitUntilDone(t, true, downloader, downloader2)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()
	downloader2.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)
	res2 := loadDataFromPersister(downloaderPersister2)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)
	os.Remove(downloaderPersister2.Path)

	checkDownloadResult(t, res, TorrentS, SeedS, output)
	checkDownloadResult(t, res2, TorrentS, SeedS, output2)

	util.EndTest()
}

func TestTwoClientsLargeFile(t *testing.T) {
	util.StartTest("Testing 36 piece file with one seeder and one downloader...")
	output := generateOutFile()
	seederPersister := makePersister()
	downloaderPersister := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)

	waitUntilDone(t, true, downloader)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	checkDownloadResult(t, res, TorrentM, SeedM, output)

	util.EndTest()
}

func TestMultipleSeedersDownloaders(t *testing.T) {
	util.StartTest("Testing 36-piece file with two seeders and two downloaders...")
	output := generateOutFile()
	output2 := generateOutFile()
	seederPersister := makePersister()
	seederPersister2 := makePersister()
	downloaderPersister := makePersister()
	downloaderPersister2 := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	seeder2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister2)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)
	downloader2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output2, downloaderPersister2)

	waitUntilDone(t, true, downloader, downloader2)

	tr.Kill()
	seeder.Kill()
	seeder2.Kill()
	downloader.Kill()
	downloader2.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)
	res2 := loadDataFromPersister(downloaderPersister2)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)
	os.Remove(downloaderPersister2.Path)

	checkDownloadResult(t, res, TorrentM, SeedM, output)
	checkDownloadResult(t, res2, TorrentM, SeedM, output2)

	util.EndTest()
}

func TestRestartedSeeder(t *testing.T) {
	util.StartTest("Testing 36-piece file with stopped seeder...")
	output := generateOutFile()
	seederPersister := makePersister()
	downloaderPersister := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)

	waitUntilStarted(t, downloader)
	seeder.Kill()
	util.Wait(1000)

	seeder = btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	if !util.AllTrue(seeder.AtomicGetBitmap()) {
		t.Fatalf("Seeder re-initialized bitmap not all true")
	}

	waitUntilDone(t, true, downloader)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	checkDownloadResult(t, res, TorrentM, SeedM, output)

	util.EndTest()
}

func TestRestartedDownloader(t *testing.T) {
	util.StartTest("Testing 36-piece file with stopped downloader...")
	output := generateOutFile()
	seederPersister := makePersister()
	downloaderPersister := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)

	waitUntilStarted(t, downloader)
	downloader.Kill()

	util.Wait(WaitForDeath) // wait for shutdown
	oldBitmap := downloader.AtomicGetBitmap()
	util.Wait(1000)

	downloader = btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)
	if !util.BoolArrayEquals(oldBitmap, downloader.AtomicGetBitmap()) {
		t.Fatalf("Downloader bitmap changed after restarting")
	}

	waitUntilDone(t, true, downloader)

	tr.Kill()
	seeder.Kill()
	downloader.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)

	checkDownloadResult(t, res, TorrentM, SeedM, output)

	util.EndTest()
}

func TestStoppedSeeder(t *testing.T) {
	util.StartTest("Testing 36-piece file with one good seeder, one stopped seeder...")
	output := generateOutFile()
	output2 := generateOutFile()
	seederPersister := makePersister()
	seederPersister2 := makePersister()
	downloaderPersister := makePersister()
	downloaderPersister2 := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	seeder2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister2)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)
	downloader2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output2, downloaderPersister2)

	waitUntilStarted(t, downloader, downloader2)
	util.Wait(50)
	seeder.Kill()

	waitUntilDone(t, true, downloader, downloader2)

	tr.Kill()
	seeder2.Kill()
	downloader.Kill()
	downloader2.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)
	res2 := loadDataFromPersister(downloaderPersister2)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)
	os.Remove(downloaderPersister2.Path)

	checkDownloadResult(t, res, TorrentM, SeedM, output)
	checkDownloadResult(t, res2, TorrentM, SeedM, output2)

	util.EndTest()
}

// check to see if seeders and downloaders can successfully finish even when tracker is taken down
func TestStoppedTracker(t *testing.T) {
	util.StartTest("Testing 36-piece file with two seeders, two downloaders, stopped tracker...")
	output := generateOutFile()
	output2 := generateOutFile()
	seederPersister := makePersister()
	seederPersister2 := makePersister()
	downloaderPersister := makePersister()
	downloaderPersister2 := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	seeder2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister2)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)
	downloader2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output2, downloaderPersister2)

	waitUntilStarted(t, downloader, downloader2)
	tr.Kill()

	waitUntilDone(t, true, downloader, downloader2)

	seeder.Kill()
	seeder2.Kill()
	downloader.Kill()
	downloader2.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)
	res2 := loadDataFromPersister(downloaderPersister2)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)
	os.Remove(downloaderPersister2.Path)
	checkDownloadResult(t, res, TorrentM, SeedM, output)
	checkDownloadResult(t, res2, TorrentM, SeedM, output2)

	util.EndTest()
}

// start one seeder and one downloader, kill the seeder when the downloader is finished,
// start up a new downloader that gets seeded, kill the first downloader when that one is done,
// start two more downloaders
func TestChain(t *testing.T) {
	util.StartTest("Testing 36-piece file with 4 chained downloaders...")
	output := generateOutFile()
	output2 := generateOutFile()
	output3 := generateOutFile()
	output4 := generateOutFile()
	seederPersister := makePersister()
	downloaderPersister := makePersister()
	downloaderPersister2 := makePersister()
	downloaderPersister3 := makePersister()
	downloaderPersister4 := makePersister()

	tr := bttracker.StartBTTracker(TorrentM, PortM)
	seeder := btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", seederPersister)
	downloader := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output, downloaderPersister)

	waitUntilDone(t, true, downloader)
	seeder.Kill()
	util.Wait(1000)
	downloader2 := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output2, downloaderPersister2)
	waitUntilDone(t, true, downloader, downloader2)
	downloader.Kill()
	util.Wait(1000)
	downloader3 := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output3, downloaderPersister3)
	downloader4 := btclient.StartBTClient("localhost", nextPort(), TorrentM, "", output4, downloaderPersister4)
	waitUntilDone(t, true, downloader, downloader2, downloader3, downloader4)

	tr.Kill()
	downloader2.Kill()
	downloader3.Kill()
	downloader4.Kill()

	util.Wait(WaitForDeath)

	res := loadDataFromPersister(downloaderPersister)
	res2 := loadDataFromPersister(downloaderPersister2)
	res3 := loadDataFromPersister(downloaderPersister3)
	res4 := loadDataFromPersister(downloaderPersister4)

	os.Remove(seederPersister.Path)
	os.Remove(downloaderPersister.Path)
	os.Remove(downloaderPersister2.Path)
	os.Remove(downloaderPersister3.Path)
	os.Remove(downloaderPersister4.Path)

	checkDownloadResult(t, res, TorrentM, SeedM, output)
	checkDownloadResult(t, res2, TorrentM, SeedM, output2)
	checkDownloadResult(t, res3, TorrentM, SeedM, output3)
	checkDownloadResult(t, res4, TorrentM, SeedM, output4)

	util.EndTest()
}

// check a healthy swarm of seeders and downloaders to see if they cooperate successfully
func TestSwarm(t *testing.T) {
	util.StartTest("Testing 36-piece file with 5 seeders, 5 downloaders...")
	numSeeders := 5
	numDownloaders := 5
	outputs := []string{}
	downloadPersisters := []*btclient.Persister{}
	seedPersisters := []*btclient.Persister{}
	downloaders := []*btclient.BTClient{}
	seeders := []*btclient.BTClient{}
	tr := bttracker.StartBTTracker(TorrentM, PortM)

	for i := 0; i < numDownloaders; i++ {
		newPersister := makePersister()
		newOutput := generateOutFile()
		outputs = append(outputs, newOutput)
		downloadPersisters = append(downloadPersisters, newPersister)
		downloaders = append(downloaders, btclient.StartBTClient("localhost", nextPort(), TorrentM, "", newOutput, newPersister))
	}
	for i := 0; i < numSeeders; i++ {
		newPersister := makePersister()
		seedPersisters = append(seedPersisters, newPersister)
		seeders = append(seeders, btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", newPersister))
	}

	waitUntilDone(t, true, downloaders...)

	tr.Kill()
	for i := 0; i < numSeeders; i++ {
		seeders[i].Kill()
		p := seedPersisters[i]
		os.Remove(p.Path)
	}
	for i := 0; i < numDownloaders; i++ {
		downloaders[i].Kill()
		p := downloadPersisters[i]
		res := loadDataFromPersister(p)
		os.Remove(p.Path)
		checkDownloadResult(t, res, TorrentM, SeedM, outputs[i])
	}

	util.EndTest()
}

// spin up a swarm of seeders and downloaders, wait for one downloader to finish first,
// kill all the seeders, and check for successful eventual downloading (with some random dropouts
// thrown in for fun)
func TestSwarmDropout(t *testing.T) {
	util.StartTest("Testing 36-piece file with 5 seeders, 5 downloaders, killing all seeders...")
	numSeeders := 5
	numDownloaders := 5
	outputs := []string{}
	downloadPersisters := []*btclient.Persister{}
	seedPersisters := []*btclient.Persister{}
	downloaders := []*btclient.BTClient{}
	seeders := []*btclient.BTClient{}
	tr := bttracker.StartBTTracker(TorrentM, PortM)

	for i := 0; i < numDownloaders; i++ {
		newPersister := makePersister()
		newOutput := generateOutFile()
		outputs = append(outputs, newOutput)
		downloadPersisters = append(downloadPersisters, newPersister)
		downloaders = append(downloaders, btclient.StartBTClient("localhost", nextPort(), TorrentM, "", newOutput, newPersister))
	}
	for i := 0; i < numSeeders; i++ {
		newPersister := makePersister()
		seedPersisters = append(seedPersisters, newPersister)
		seeders = append(seeders, btclient.StartBTClient("localhost", nextPort(), TorrentM, SeedM, "", newPersister))
	}

	waitUntilStarted(t, downloaders...)
	seeders[0].Kill()                 // kill one of the seeders for fun
	kill := rand.Intn(numDownloaders) // kill one of the downloaders for fun
	downloaders[kill].Kill()
	util.Wait(1000)
	downloaders[kill] = btclient.StartBTClient("localhost", nextPort(), TorrentM, "", outputs[kill], downloadPersisters[kill])

	waitUntilDone(t, false, downloaders...) // wait until one of the downloaders is done
	for i := 1; i < numSeeders; i++ {       // kill all remaining seeders
		seeders[i].Kill()
	}
	kill = rand.Intn(numDownloaders) // kill one of the downloaders for fun
	downloaders[kill].Kill()
	util.Wait(1000)
	downloaders[kill] = btclient.StartBTClient("localhost", nextPort(), TorrentM, "", outputs[kill], downloadPersisters[kill])

	waitUntilDone(t, true, downloaders...)

	tr.Kill()
	for i := 0; i < numSeeders; i++ {
		seeders[i].Kill()
		p := seedPersisters[i]
		os.Remove(p.Path)
	}
	for i := 0; i < numDownloaders; i++ {
		downloaders[i].Kill()
		p := downloadPersisters[i]
		res := loadDataFromPersister(p)
		os.Remove(p.Path)
		checkDownloadResult(t, res, TorrentM, SeedM, outputs[i])
	}

	util.EndTest()
}
