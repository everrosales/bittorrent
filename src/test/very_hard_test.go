package test

import (
	"client"
	"math/rand"
	"os"
	"testing"
	"tracker"
	"util"
)

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
