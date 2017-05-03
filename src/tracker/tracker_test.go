package bttracker

import (
	"fs"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"util"
)

type requestParams struct {
	peerId     string
	ip         string
	port       string
	uploaded   int
	downloaded int
	left       int
	infoHash   string
}

const BaseStr = "http://localhost:8000/"
const Peer1 = "aaaaaaaaaaaaaaaaaaaa"
const Peer2 = "bbbbbbbbbbbbbbbbbbbb"
const Port1 = "6882"
const Port2 = "6883"
const InfoHash = "oRoSncDAfu3-cCmwPVg9CalY8_o="

// Helpers
func makeTestTracker(port int) *BTTracker {
	util.Debug = util.Trace
	return StartBTTracker("../main/test.torrent", port)
}

func makeUrlWithParams(req *requestParams) string {
	result := BaseStr + "?peer_id=" + req.peerId + "&port=" + req.port +
		"&ip=" + req.ip + "&uploaded=" + strconv.Itoa(req.uploaded) +
		"&downloaded=" + strconv.Itoa(req.downloaded) + "&left=" +
		strconv.Itoa(req.left) + "&info_hash=" + req.infoHash
	return result
}

// Tests
func TestMakeTracker(t *testing.T) {
	util.StartTest("Testing basic starting and killing of tracker...")
	tr := makeTestTracker(8000)
	tr.Kill()
	util.EndTest()
}

func TestBasicRequest(t *testing.T) {
	util.StartTest("Testing basic request to tracker...")
	tr := makeTestTracker(8001)
	params := requestParams{Peer1, "", Port1, 0, 0, 300, InfoHash}
	url := makeUrlWithParams(&params)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Error sending request")
	}
	if resp.Status != "200 OK" {
		t.Fatalf("Wrong response status code")
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		t.Fatalf("Failure reading response body")
	}
	respF := FailureResponse{}
	fs.Decode(string(bodyBytes), &respF)
	if respF.Failure != "" {
		t.Fatalf("Received failure response from server: %s", respF.Failure)
	}
	respS := SuccessResponse{}
	fs.Decode(string(bodyBytes), &respS)
	me := respS.Peers[0]
	if me["peer id"] != Peer1 {
		t.Fatalf("Wrong peer id")
	}
	if me["port"] != Port1 {
		t.Fatalf("Missing port")
	}
	if me["ip"] != "::1" {
		t.Fatalf("Missing ip")
	}
	tr.Kill()
	util.EndTest()
}

// TODO: add tests for expected failure (e.g. missing param, non-int param)

// TODO: add tests for peer logic (multiple peers)

// TODO: add tests for many peers (like 100)

// TODO: add tests for peer heartbeats
