package bttracker

import (
	"errors"
	"fs"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

const BaseStr = "http://localhost:"
const Peer1 = "aaaaaaaaaaaaaaaaaaaa"
const Peer2 = "bbbbbbbbbbbbbbbbbbbb"
const Port1 = "6882"
const Port2 = "6883"
const InfoHash = "HmbK7rlK8tBmNJtShTaW23s-H_Q="

func init() {
	util.Debug = util.None
}

// Helpers
func makeTestTracker(port int) *BTTracker {
	return StartBTTracker("../main/test.torrent", port)
}

func sendRequest(port int, req *requestParams) ([]byte, error) {
	url := BaseStr + strconv.Itoa(port) + "/?peer_id=" + req.peerId +
		"&port=" + req.port + "&ip=" + req.ip + "&uploaded=" +
		strconv.Itoa(req.uploaded) + "&downloaded=" + strconv.Itoa(req.downloaded) +
		"&left=" + strconv.Itoa(req.left) + "&info_hash=" + req.infoHash
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Error sending request")
	}
	if resp.Status != "200 OK" {
		return nil, errors.New("Wrong response status code")
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, errors.New("Failure reading response body")
	}
	return bodyBytes, nil
}

// Tests
func TestMakeTracker(t *testing.T) {
	util.StartTest("Testing basic starting and killing of tracker...")
	tr := makeTestTracker(8000)
	tr.Kill()
	util.Wait(500)
	util.EndTest()
}

func TestBasicRequest(t *testing.T) {
	util.StartTest("Testing basic request to tracker...")
	tr := makeTestTracker(8001)
	params := requestParams{Peer1, "", Port1, 0, 0, 300, InfoHash}
	bodyBytes, err := sendRequest(8001, &params)
	if err != nil {
		tr.Kill()
		t.Fatalf("%s\n", err.Error())
	}
	respF := FailureResponse{}
	fs.Decode(string(bodyBytes), &respF)
	if respF.Failure != "" {
		tr.Kill()
		t.Fatalf("Received failure response from server: %s", respF.Failure)
	}
	respS := SuccessResponse{}
	fs.Decode(string(bodyBytes), &respS)
	me := respS.Peers[0]
	if me["peer id"] != Peer1 {
		tr.Kill()
		t.Fatalf("Wrong peer id")
	}
	if me["port"] != Port1 {
		tr.Kill()
		t.Fatalf("Missing port")
	}
	if me["ip"] != "::1" {
		tr.Kill()
		t.Fatalf("Missing ip")
	}
	tr.Kill()
	util.Wait(500)
	util.EndTest()
}

func TestBadInfoHashRequest(t *testing.T) {
	util.StartTest("Testing expected failure for request with bad infohash...")
	tr := makeTestTracker(8002)
	params := requestParams{Peer1, "", Port1, 0, 0, 300, "r39n9S3lzSQTBj5Jk6JBfdVfYu0="}
	bodyBytes, err := sendRequest(8002, &params)
	if err != nil {
		tr.Kill()
		t.Fatalf("%s\n", err.Error())
	}
	respF := FailureResponse{}
	fs.Decode(string(bodyBytes), &respF)
	if !strings.Contains(respF.Failure, "invalid infohash") {
		tr.Kill()
		t.Fatalf("Expected invalid infohash response from server")
	}
	tr.Kill()
	util.Wait(500)
	util.EndTest()
}

func TestBadPortRequest(t *testing.T) {
	util.StartTest("Testing expected failure for request with bad port...")
	tr := makeTestTracker(8003)
	params := requestParams{Peer1, "", "port", 0, 0, 300, InfoHash}
	bodyBytes, err := sendRequest(8003, &params)
	if err != nil {
		tr.Kill()
		t.Fatalf("%s\n", err.Error())
	}
	respF := FailureResponse{}
	fs.Decode(string(bodyBytes), &respF)
	if !strings.Contains(respF.Failure, "bad parameter") {
		tr.Kill()
		t.Fatalf("Expected bad parameter response from server")
	}
	tr.Kill()
	util.Wait(500)
	util.EndTest()
}

func TestBadPeerIdRequest(t *testing.T) {
	util.StartTest("Testing expected failure for request with bad peer id...")
	tr := makeTestTracker(8004)
	params := requestParams{"peerid", "", Port1, 0, 0, 300, InfoHash}
	bodyBytes, err := sendRequest(8004, &params)
	if err != nil {
		tr.Kill()
		t.Fatalf("%s\n", err.Error())
	}
	respF := FailureResponse{}
	fs.Decode(string(bodyBytes), &respF)
	if !strings.Contains(respF.Failure, "invalid peerId") {
		tr.Kill()
		t.Fatalf("Expected invalid peerId response from server")
	}
	tr.Kill()
	util.Wait(500)
	util.EndTest()
}

// TODO: add tests for peer logic (multiple peers)

// TODO: add tests for many peers (like 100)

// TODO: add tests for peer heartbeats
