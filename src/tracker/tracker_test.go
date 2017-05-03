package bttracker

import (
	"errors"
	"fmt"
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
const BetweenTests = 50

func init() {
	util.Debug = util.None
}

// Helpers
func makeTestTracker(port int) *BTTracker {
	return StartBTTracker("../main/test.torrent", port)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func findPeer(peerId string, peers []map[string]string) (map[string]string, error) {
	for _, e := range peers {
		if e["peer id"] == peerId {
			return e, nil
		}
	}
	return make(map[string]string), errors.New("can't find peer")
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
	util.Wait(BetweenTests)
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
	util.Wait(BetweenTests)
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
	util.Wait(BetweenTests)
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
	util.Wait(BetweenTests)
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
	util.Wait(BetweenTests)
	util.EndTest()
}

func TestMultiplePeersBasic(t *testing.T) {
	util.StartTest("Testing multiple peers basic...")
	tr := makeTestTracker(8005)
	params := requestParams{Peer1, "", Port1, 0, 0, 300, InfoHash}
	bodyBytes, err := sendRequest(8005, &params)
	if err != nil {
		tr.Kill()
		t.Fatalf("%s\n", err.Error())
	}
	respS := SuccessResponse{}
	fs.Decode(string(bodyBytes), &respS)
	_, err = findPeer(Peer1, respS.Peers)
	if len(respS.Peers) != 1 || err != nil {
		tr.Kill()
		t.Fatalf("Expected 1 peer, got %d\n", len(respS.Peers))
	}

	params = requestParams{Peer2, "", Port2, 0, 0, 300, InfoHash}
	bodyBytes, err = sendRequest(8005, &params)
	if err != nil {
		tr.Kill()
		t.Fatalf("%s\n", err.Error())
	}
	respS = SuccessResponse{}
	fs.Decode(string(bodyBytes), &respS)
	_, err1 := findPeer(Peer1, respS.Peers)
	_, err2 := findPeer(Peer2, respS.Peers)
	if len(respS.Peers) != 2 || err1 != nil || err2 != nil {
		tr.Kill()
		t.Fatalf("Expected 2 peer, got %d\n", len(respS.Peers))
	}

	tr.Kill()
	util.Wait(BetweenTests)
	util.EndTest()
}

func TestManyPeers(t *testing.T) {
	util.StartTest("Testing many peers...")
	tr := makeTestTracker(8006)

	var params requestParams
	var respS SuccessResponse
	var bodyBytes []byte
	var err error
	baseStr := "aaaaaaaaaaaaaaa"
	for i := 0; i < 50; i++ {
		peerId := baseStr + fmt.Sprintf("%05d", i)
		params = requestParams{peerId, "", Port1, 0, 0, 300, InfoHash}
		bodyBytes, err = sendRequest(8006, &params)
		if err != nil {
			tr.Kill()
			t.Fatalf("%s\n", err.Error())
		}
		respS = SuccessResponse{}
		fs.Decode(string(bodyBytes), &respS)
		_, err = findPeer(peerId, respS.Peers)
		if len(respS.Peers) != i+1 || err != nil {
			tr.Kill()
			t.Fatalf("Expected %d peers, got %d\n", i+1, len(respS.Peers))
		}
	}
	for i := 0; i < 50; i++ {
		peerId := baseStr + fmt.Sprintf("%05d", i)
		params = requestParams{peerId, "", Port1, 0, 0, 300, InfoHash}
		bodyBytes, err = sendRequest(8006, &params)
		if err != nil {
			tr.Kill()
			t.Fatalf("%s\n", err.Error())
		}
		respS = SuccessResponse{}
		fs.Decode(string(bodyBytes), &respS)
		_, err = findPeer(peerId, respS.Peers)
		if len(respS.Peers) != 50 {
			tr.Kill()
			t.Fatalf("Expected %d peers, got %d\n", 50, len(respS.Peers))
		}
	}

	tr.Kill()
	util.Wait(BetweenTests)
	util.EndTest()
}

// TODO: add tests for peer heartbeats
