package fs

import (
	"net/url"
	"os"
	"testing"
	"util"
)

const TempTorrent = "temp.torrent"

func init() {
	util.Debug = util.None
}

func TestReadRealTorrent(t *testing.T) {
	util.StartTest("Testing reading a real torrent...")
	metadata := Read("../main/test.torrent")
	if metadata.TrackerUrl != "http://tracker.raspberrypi.org:6969/announce" {
		t.Fatalf("Torrent URL unexpected")
	}
	util.EndTest()
}

func TestInfoHash(t *testing.T) {
	util.StartTest("Testing info hash reading and encoding...")
	torrent := ReadTorrent("../main/test.torrent")
	infoHash := GetInfoHash(torrent)
	if url.QueryEscape(infoHash) != "%C7%A2%CEDd0A%7B%FE%16%82%5C%BDa%DD6%DD%1DS%C1" {
		t.Fatalf("Decoding info hash failed")
	}
	util.EndTest()
}

func TestReadFakeTorrent(t *testing.T) {
	util.StartTest("Testing writing and reading a fake torrent...")
	file := FileData{Length: 1234}
	Write(TempTorrent, Metadata{"blahUrl", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []FileData{file}})

	torrent := ReadTorrent(TempTorrent)
	if torrent.Announce != "blahUrl" {
		t.Fatalf("Torrent URLs don't match")
	}

	metadata := Read(TempTorrent)
	if metadata.TrackerUrl != "blahUrl" {
		t.Fatalf("Torrent URLs don't match")
	}

	err := os.Remove(TempTorrent)
	if err != nil {
		util.EPrintf("Failed to delete temp torrent\n")
	}

	util.EndTest()
}
