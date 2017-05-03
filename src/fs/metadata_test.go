package fs

import "testing"
import "util"

func TestReadFakeTorrent(t *testing.T) {
    util.TestStartPrintf("TestReadFakeTorrent")
    file := FileData{Length: 1234}
    Write("test.torrent", Metadata{"blah", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []FileData{file}})
    Read("test.torrent")
    util.TestFinishPrintf("Passed")
}

func TestReadRealTorrent(t *testing.T) {
    util.TestStartPrintf("TestReadRealTorrent")
    Read("../main/test.torrent")
    util.TestFinishPrintf("Passed")
}