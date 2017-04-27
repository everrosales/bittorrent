package main

import (
	//"client"
	"flag"
	"tracker"
	"util"
)

func main() {
	clientFlag := flag.Bool("client", false, "Start client for torrent")
	trackerFlag := flag.Bool("tracker", false, "Start tracker for torrent")
	fileFlag := flag.String("file", "", "Torrent file (required)")
	debugFlag := flag.String("debug", "None", "Debug level [None|Info|Trace]")
	portFlag := flag.Int("port", 8000, "Port (default 8000)")
	flag.Parse()

	// set debug level
	if *debugFlag == "None" {
		util.Debug = util.None
	} else if *debugFlag == "Info" {
		util.Debug = util.Info
	} else if *debugFlag == "Trace" {
		util.Debug = util.Trace
	} else {
		util.EPrintf("Invalid debug level.\n")
		return
	}

	// check for file flag, since it's required
	if *fileFlag == "" {
		util.EPrintf("Missing file flag.\n")
		return
	}

	// check for valid port
	if *portFlag < 1 || *portFlag > 65535 {
		util.EPrintf("Invalid port number.\n")
		return
	}

	// start client or tracker
	if *clientFlag == *trackerFlag {
		util.EPrintf("Select either client or tracker.\n")
		return
	} else if *trackerFlag {
		bttracker.StartBTTracker(*fileFlag, *portFlag)
	} else if *clientFlag {
		// TODO: start client
	}

}
