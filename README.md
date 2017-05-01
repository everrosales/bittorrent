# 6.824 Final Project
For our final project for MIT's 6.824 (Distributed Systems), our team implemented a Bittorrent client and tracker in Go. 

## Team Members
* Nicole Glabinski (nicolekg)
* Everardo Rosales (erosales)
* Kimberli Zhong (kimberli)

## Requirements
Requires Go `1.8` or greater, for shutdown of HTTP servers.

You'll also need to install http://github.com/zeebo/bencode/. In the `src` folder, run `go get github.com/zeebo/bencode`.

## Usage
To run either the tracker or client, go in to `src/main` and run `go run main.go`. Run the tracker with flag `-tracker` and run the client with flag `-client`. Specify the `.torrent` file you want to use with `-torrent=<NAME>`. Other flags include `-debug` and `-port`. 

## Development
* `src/client` - code for the client
* `src/tracker` - code for the tracker
* `src/fs` - 
* `src/btnet` - 
* `src/github.com` - 
* `src/util` - utils for development (e.g. debug printing)
* `src/main` - the main command-line utility
