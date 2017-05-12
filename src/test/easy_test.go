package test

import (
	"client"
	"os"
	"testing"
	"tracker"
	"util"
)

func init() {
	os.MkdirAll("out", 0777)
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
