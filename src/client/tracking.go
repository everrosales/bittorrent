package btclient

import (
	"btnet"
	"errors"
	"fs"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"util"
)

type trackerReq struct {
	peerId     string
	ip         string
	port       string
	uploaded   int
	downloaded int
	left       int
	infoHash   string
	status     status
}

type TrackerRes struct {
	Interval int                 `bencode:"interval"`
	Peers    []map[string]string `bencode:"peers"`
	Failure  string              `bencode:"failure reason"`
}

func (cl *BTClient) trackerHeartbeat() {
	for {
		if cl.CheckShutdown() {
			return
		}
		res := cl.contactTracker(cl.torrentMeta.TrackerUrl)
        for _, p := range res.Peers {
			util.TPrintf("%s: peerId %s, ip %s, port %s\n", cl.port, p["peer id"], p["ip"], p["port"])
			addr, err := net.ResolveTCPAddr("tcp", p["ip"]+":"+p["port"])
			if err != nil {
				// panic(err)
                continue
			}
			myAddr, err := net.ResolveTCPAddr("tcp", cl.ip+":"+cl.port)
			if addr.String() != myAddr.String() {
				util.TPrintf("%s: sending initial message to %v\n", cl.port, addr)
				cl.SendPeerMessage(addr, btnet.PeerMessage{KeepAlive: true})
			}
		}
		cl.lock("tracking/trackerHeartbeat")
		wait := cl.heartbeatInterval * 1000
        cl.unlock("tracking/trackerHeartbeat")
		util.Wait(wait)
	}
}

func (cl *BTClient) contactTracker(baseUrl string) TrackerRes {
	// TODO: update uploaded, downloaded, and left
	cl.lock("tracking/contactTracker 1")
	request := trackerReq{cl.peerId, cl.ip, cl.port, 0, 0, 0, cl.infoHash, cl.status}
	cl.unlock("tracking/contactTracker 1")
	byteRes, err := sendRequest(baseUrl, &request)
	if err != nil {
		util.WPrintf("Received error sending to tracker: %s\n", err)
	}
	res := TrackerRes{}
	fs.Decode(byteRes, &res)
	if res.Failure != "" {
		util.WPrintf("Received error from tracker: %s\n", res.Failure)
	}
	for _, p := range res.Peers {
		if _, ok := p["port"]; !ok {
			util.WPrintf("bad peers response: %v\n", res.Peers)
			panic("got port 0 from tracker")
		}
	}
	util.TPrintf("Contacting tracker at %s (%d peers)\n", baseUrl, len(res.Peers))
	cl.lock("tracking/contactTracker 2")
	cl.heartbeatInterval = res.Interval
	cl.unlock("tracking/contactTracker 2")
	return res
}

func sendRequest(addr string, req *trackerReq) ([]byte, error) {
	url := addr + "/?peer_id=" + url.QueryEscape(req.peerId) +
		"&port=" + req.port + "&ip=" + req.ip + "&uploaded=" +
		strconv.Itoa(req.uploaded) + "&downloaded=" + strconv.Itoa(req.downloaded) +
		"&left=" + strconv.Itoa(req.left) + "&info_hash=" +
		url.QueryEscape(req.infoHash) + "&status=" + string(req.status)
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
