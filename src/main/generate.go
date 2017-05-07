package main


import (
	"flag"
	"util"
	"fs"
	"strings"
)

func xmain() {
	fileFlag := flag.String("file", "", "Input file path (required)")
	outFlag := flag.String("out", "", "Output torrent file name (default is 'out.torrent')")
	urlFlag := flag.String("url", "", "URL of tracker (required)")
	nameFlag := flag.String("name", "", "Suggested name to save file as (required)")
	flag.Parse()

	torrentFile := *outFlag
	filePath := *fileFlag
	url := *urlFlag
	name := *nameFlag

	if filePath == "" {
		util.EPrintf("Missing input file path (-file)\n")
		return
	}
	if url == "" {
		util.EPrintf("Missing tracker url (-url)\n")
		return
	}
	if name == "" {
		util.EPrintf("Missing torrent name (-name)\n")
		return
	}
	if torrentFile == "" || !strings.Contains(torrentFile, ".torrent") {
		util.IPrintf("Missing torrent file name, writing to 'out.torrent'\n")
		torrentFile = "out.torrent"
	}

	metadata := fs.GetMetadata(filePath, url, name, 10)
	fs.Write(torrentFile, metadata)
}
