package bttracker

import (
	"fmt"
	"fs"
	"net/http"
	"strconv"
	"util"
)

const (
	PeerIdLength = 20
)

// custom app handler with parameterized context
type appHandler struct {
	trackerContext *BTTracker
	H              func(*BTTracker, http.ResponseWriter, *http.Request) (int, error)
}

type SuccessResponse struct {
	Interval int
	Peers    []map[string]string
}

type FailureResponse struct {
	Failure string
}

func writeSuccess(w http.ResponseWriter, interval int, peers []map[string]string) (int, error) {
	// TODO: escape response
	resp := fs.Encode(SuccessResponse{interval, peers})
	fmt.Fprintf(w, resp)
	return 200, nil
}

func writeFailure(w http.ResponseWriter, format string, a ...interface{}) (int, error) {
	// TODO: escape response
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
	// parsing query params
	infoHash := r.URL.Query().Get("info_hash")
	peerId := r.URL.Query().Get("peer_id")
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
	} else if len(peerId) != PeerIdLength {
		return writeFailure(w, "invalid peerId %s", peerId)
	} else if event != Started && event != Completed && event != Stopped && event != Empty {
		return writeFailure(w, "invalid event %s", event)
	}

	// good request; applying update
	peer := peer{peerId, ip, port, uploaded, downloaded, left, event}

	tr.mu.Lock()
	tr.peers[peerId] = peer
	numPeers := len(tr.peers)
	peers := tr.getPeers()
	tr.mu.Unlock()

	util.TPrintf("Received request from %s, now have %d peer(s)\n", peerId, numPeers)
	return writeSuccess(w, 0, peers)
}

func (tr *BTTracker) main(port int) {
	util.IPrintf("Tracker for %s listening on port %d\n", tr.file, port)
	http.Get("/")
	http.ListenAndServe(":"+strconv.Itoa(port), appHandler{tr, IndexHandler})
}
