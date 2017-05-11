package main

import (
	"client"
	"flag"
	"io/ioutil"
	"os"
	"tracker"
	"util"
)

func main() {
	// TODO: add utility for making a .torrent file
	clientFlag := flag.Bool("client", false, "Start client for torrent")
	trackerFlag := flag.Bool("tracker", false, "Start tracker for torrent")
	fileFlag := flag.String("file", "", "Torrent (.torrent) file (required)")
	seedFlag := flag.String("seed", "", "The file for the client to seed (client only)")
	seedFlag := flag.String("output", "", "The path to save the downloaded file (client only)")
	debugFlag := flag.String("debug", "None", "Debug level [None|Info|Trace|Lock]")
	portFlag := flag.Int("port", 8000, "Port (default 8000)")
	flag.Parse()

	// set debug level
	if *debugFlag == "None" {
		util.Debug = util.None
	} else if *debugFlag == "Info" {
		util.Debug = util.Info
	} else if *debugFlag == "Trace" {
		util.Debug = util.Trace
	} else if *debugFlag == "Lock" {
		util.Debug = util.Lock
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
		if *seedFlag != "" {
			util.EPrintf("Trackers cannot seed files.\n")
			return
		}
		tr := bttracker.StartBTTracker(*fileFlag, *portFlag)
		for !tr.CheckShutdown() {
		}
		return
	} else if *clientFlag {
		// StartBTClient(ip string, port int, metadataPath string, persister *Persister)
		tmpFile, err := ioutil.TempFile(".", *fileFlag+"_download")
		if err != nil {
			panic(err)
		}

		cl := btclient.StartBTClient("localhost", *portFlag, *fileFlag, *seedFlag, *outputFlag, btclient.MakePersister(tmpFile.Name()))
		for !cl.CheckShutdown() {
		}
		os.Remove(tmpFile.Name())
		return
	}

}
