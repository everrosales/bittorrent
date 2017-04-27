package bttracker

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Peer struct {
	peer_id    string
	ip         string
	port       int
	uploaded   int
	downloaded int
}

type BTTracker struct {
	file      string
	info_hash string
	mu        sync.Mutex
	peers     []Peer
	count     int
}

func StartBTTracker(file string) *BTTracker {
	tr := &BTTracker{}
	tr.file = file
	// read info_hash from file
	go tr.main()
	return tr
}

// custom app handler with parameterized context
type appHandler struct {
	trackerContext *BTTracker
	H              func(*BTTracker, http.ResponseWriter, *http.Request) (int, error)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := ah.H(ah.trackerContext, w, r)
	if err != nil {
		log.Printf("HTTP %d: %q", status, err)
		http.Error(w, http.StatusText(status), status)
	}
}

// handle GET /
func IndexHandler(a *BTTracker, w http.ResponseWriter, r *http.Request) (int, error) {
	a.count += 1

	a.info_hash = r.URL.Query().Get("info_hash")
	//peer_id := r.URL.Query().Get("peer_id")
	//ip := r.URL.Query().Get("ip")
	//port := r.URL.Query().Get("port")
	//uploaded := r.URL.Query().Get("uploaded")
	//downloaded := r.URL.Query().Get("downloaded")
	//left := r.URL.Query().Get("left")
	//event := r.URL.Query().Get("event")

	fmt.Fprintf(w, "hi user: %d, info_hash: %s", a.count, a.info_hash)
	fmt.Printf("new count: %d\n", a.count)
	return 200, nil
}

func (bt *BTTracker) main() {
	http.Get("/")
	http.ListenAndServe(":8000", appHandler{bt, IndexHandler})
}
