package bttracker

import (
	"fmt"
	"fs"
	"net"
	"net/http"
	"strconv"
	"time"
	"util"
)

const (
	PeerIdLength    = 20
	DefaultInterval = 5 // seconds; sent to clients in response
)

// custom app handler with parameterized context
type appHandler struct {
	trackerContext *BTTracker
	H              func(*BTTracker, http.ResponseWriter, *http.Request) (int, error)
}

type SuccessResponse struct {
	Interval int                 `bencode:"interval"`
	Peers    []map[string]string `bencode:"peers"`
}

type FailureResponse struct {
	Failure string `bencode:"failure reason"`
}

func writeSuccess(w http.ResponseWriter, interval int, peers []map[string]string) (int, error) {
	resp := fs.Encode(SuccessResponse{interval, peers})
	util.TPrintf("[resp] %v\n", resp)
	fmt.Fprintf(w, resp)
	return 200, nil
}

func writeFailure(w http.ResponseWriter, format string, a ...interface{}) (int, error) {
	resp := fs.Encode(FailureResponse{fmt.Sprintf(format, a...)})
	fmt.Fprintf(w, resp)
	return 200, nil
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	status, err := ah.H(ah.trackerContext, w, r)
	if err != nil {
		util.EPrintf("HTTP %d: %q\n", status, err)
		http.Error(w, err.Error(), status)
	}
}

// handle GET /
func IndexHandler(tr *BTTracker, w http.ResponseWriter, r *http.Request) (int, error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// parsing query params
	infoHash := r.URL.Query().Get("info_hash")
	peerIdStr := peerId(r.URL.Query().Get("peer_id"))
	ip := r.URL.Query().Get("ip")
	port, errPort := strconv.Atoi(r.URL.Query().Get("port"))
	uploaded, errUp := strconv.Atoi(r.URL.Query().Get("uploaded"))
	downloaded, errDown := strconv.Atoi(r.URL.Query().Get("downloaded"))
	left, errLeft := strconv.Atoi(r.URL.Query().Get("left"))
	event := peerStatus(r.URL.Query().Get("event"))
	if event == peerStatus("") {
		event = Empty
	}

	// checking valid parameters
	if errPort != nil || errUp != nil || errDown != nil || errLeft != nil {
		return writeFailure(w, "bad parameter (non-integer value)")
	} else if infoHash != tr.infoHash {
		return writeFailure(w, "invalid infohash %s", infoHash)
	} else if port < 1 || port > 65535 {
		return writeFailure(w, "invalid port %s", port)
	} else if len(peerIdStr) != PeerIdLength {
		return writeFailure(w, "invalid peerId %s", peerIdStr)
	} else if event != Started && event != Completed && event != Stopped && event != Empty {
		return writeFailure(w, "invalid event %s", event)
	} else if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	// good request; applying update
	reqTime := time.Now()
	peer := peer{peerIdStr, ip, port, uploaded, downloaded, left, event, reqTime}

	tr.mu.Lock()
	tr.peers[peerIdStr] = peer
	numPeers := len(tr.peers)
	peers := tr.getPeerList()
	tr.mu.Unlock()

	util.TPrintf("[%s] req %s (ip: %s:%d), %d peer(s): %v\n", reqTime.Format("2006-01-02 15:04:05.9999"), peerIdStr, ip, port, numPeers, peers)
	return writeSuccess(w, DefaultInterval, peers)
}

func (tr *BTTracker) main(port int) {
	portStr := ":" + strconv.Itoa(port)

	tr.mu.Lock()
	tr.srv = &http.Server{Addr: portStr}
	tr.mu.Unlock()

	http.Get("/")
	go func() {
		for {
			if tr.CheckShutdown() {
				tr.srv.Close()
				util.IPrintf("Shutting down tracker on port %d...\n", tr.port)
				return
			}
			util.Wait(10)
		}
	}()
	http.ListenAndServe(portStr, appHandler{tr, IndexHandler})
}
