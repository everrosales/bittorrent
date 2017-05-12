package test

import (
	"client"
	"os"
	"testing"
	"tracker"
	"util"
)

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
