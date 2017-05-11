package main

import (
	"client"
	"flag"
	"fs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"tracker"
	"util"
)

func cleanup(cl *btclient.BTClient) {
	cl.Kill()
	// fmt.Println("cleanup")
}

func generate(input string, output string, url string, name string) {
	metadata := fs.GetMetadata(input, url, name)
	fs.Write(output, metadata)
}

func main() {
	showStatus := false
	clientFlag := flag.Bool("client", false, "Start client for torrent")
	trackerFlag := flag.Bool("tracker", false, "Start tracker for torrent")
	generateFlag := flag.Bool("generate", false, "Generate torrent file")
	torrentFlag := flag.String("torrent", "", "Torrent (.torrent) file (required)")
	seedFlag := flag.String("seed", "", "The file for the client to seed (-client only)")
	ipFlag := flag.String("ip", "localhost", "Client's IP address (default 'localhost')")
	fileFlag := flag.String("file", "", "The path to read from or write to (-client and -generate only)")
	debugFlag := flag.String("debug", "None", "Debug level [Status|None|Info|Trace|Lock]")
	urlFlag := flag.String("url", "", "URL of tracker (-generate only)")
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
	} else if *debugFlag == "Status" {
		showStatus = true
		util.Debug = util.None
	} else {
		util.EPrintf("Invalid debug level.\n")
		return
	}

	// check for file flag, since it's required
	if *torrentFlag == "" {
		util.EPrintf("Missing torrent file flag (-torrent)\n")
		return
	}

	// check for valid port
	if *portFlag < 1 || *portFlag > 65535 {
		util.EPrintf("Invalid port number\n")
		return
	}

	// start client or tracker
	if *generateFlag {
		if *fileFlag == "" {
			util.EPrintf("Need to specify what file you're trying to torrent with -file\n")
			return
		}
		if *urlFlag == "" {
			util.EPrintf("Need to specify URL of tracker with -url\n")
			return
		}
		util.Printf("Generating torrent for file %s and tracker url %s...\nSaving to %s\n", *fileFlag, *urlFlag, *torrentFlag)
		generate(*fileFlag, *torrentFlag, *urlFlag, *fileFlag)
	} else if *clientFlag == *trackerFlag {
		util.EPrintf("Select either client or tracker.\n")
		return
	} else if *trackerFlag {
		if *seedFlag != "" {
			util.EPrintf("Trackers cannot seed files.\n")
			return
		}
		tr := bttracker.StartBTTracker(*torrentFlag, *portFlag)
		for !tr.CheckShutdown() {
		}
		return
	} else if *clientFlag {
		tmpFile, err := ioutil.TempFile(".", *fileFlag+"_download")
		if err != nil {
			panic(err)
		}

		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		cl := btclient.StartBTClient(*ipFlag, *portFlag, *torrentFlag, *seedFlag, *fileFlag, btclient.MakePersister(tmpFile.Name()))

		go func() {
			<-c
			cleanup(cl)
			os.Remove(tmpFile.Name())
			os.Exit(1)
		}()

		if showStatus {
			status, _ := cl.GetStatusString()
			util.ZeroCursor()
			util.ClearScreen()
			util.Printf(status)
			for !cl.CheckShutdown() {
				util.ZeroCursor()
				status, _ = cl.GetStatusString()
				util.Printf(status)
				util.Wait(100)
			}
		} else {
			for !cl.CheckShutdown() {
			}
		}
		return
	}

}
