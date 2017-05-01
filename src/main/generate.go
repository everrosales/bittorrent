package main

import (
	"flag"
	"fs"
	"os"
	"strings"
	"util"
)

func main() {
	fileFlag := flag.String("file", "", "Input file path (required)")
	outFlag := flag.String("out", "", "Output torrent file name (default is 'out.torrent')")
	flag.Parse()

	torrentFile := *outFlag
	filePath := *fileFlag

	if filePath == "" {
		util.EPrintf("Missing input file path.\n")
		return
	}
	if torrentFile == "" || !strings.Contains(torrentFile, ".torrent") {
		util.IPrintf("Missing torrent file name, writing to 'out.torrent'\n")
		torrentFile = "out.torrent"
	}

	fi, e := os.Stat(filePath)
	if e != nil {
		util.EPrintf("Error opening file, quitting\n")
	}

	files := []string{}
	files[0] = filePath

	fileInfo := fs.FileData{fi.Size(), files}

	fs.Write(torrentFile, fs.Metadata{"blah", "blah", 1, []string{"aaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbb"}, []fs.FileData{fileInfo}})
}
