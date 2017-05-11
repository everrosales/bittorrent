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

## Testing
### Automated Testing
Run tests with `go test` in the following directories:
* `src/btnet`
* `src/client`
* `src/fs`
* `src/tracker`

You can also run integration tests in `main` by running the command `go test main_test.go`.

The file `main/torrent/test.torrent` is used in our unit tests and was downloaded from `https://downloads.raspberrypi.org/raspbian_lite_latest.torrent` in May 2017. Other test torrents used are in `main/torrent`.

### Manual Testing
Run manual tests with the following commands:
* `go run main.go -tracker -port=8000 -torrent=torrent/puppy.torrent -debug=Info`
* `go run main.go -client -port=8001 -torrent=torrent/puppy.torrent -file=out.jpg`
* `go run main.go -client -port=8002 -torrent=torrent/puppy.torrent -seed=seed/puppy.jpg`
