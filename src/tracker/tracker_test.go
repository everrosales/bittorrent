package bttracker

// tracker test
// TODO: add tests

import (
	"fs"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"
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

const Port = 8000
const BaseStr = "http://localhost:8000/"
const Peer1 = "aaaaaaaaaaaaaaaaaaaa"
const Peer2 = "bbbbbbbbbbbbbbbbbbbb"
const Port1 = "6882"
const Port2 = "6883"
const InfoHash = "fjm2CfSuEB9m5HqNDp4ZBMi-_FE="

// Helpers
func makeTestTracker() *BTTracker {
	util.Debug = util.Trace
	return StartBTTracker("../main/test.torrent", Port)
}

func startTest(desc string) *BTTracker {
	util.DPrintf(util.Default, desc)
	tr := makeTestTracker()
	return tr
}

func endTest(tr *BTTracker) {
	<-time.After(time.Millisecond * 1000)
	tr.Kill()
	util.DPrintf(util.Green, " pass\n")
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
	tr := startTest("Testing basic starting and killing of tracker...")
	endTest(tr)
}

func TestBasicRequest(t *testing.T) {
	tr := startTest("Testing basic request to tracker...")
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
	endTest(tr)
}
