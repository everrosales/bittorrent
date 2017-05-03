package fs

import (
	"testing"
	"util"
)

func init() {
	util.Debug = util.None
}

func TestReadFakeTorrent(t *testing.T) {
	util.StartTest("TestReadFakeTorrent")
	file := FileData{Length: 1234}
	Write("test.torrent", Metadata{"blah", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []FileData{file}})
	Read("test.torrent")
	util.EndTest()
}

func TestReadRealTorrent(t *testing.T) {
	util.StartTest("TestReadRealTorrent")
	Read("../main/test.torrent")
	util.EndTest()
}
