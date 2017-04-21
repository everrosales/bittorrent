package btclient
//Client

import "sync"
import "http"

type TorrentMetadata struct {
    path string
}

type BTClient struct {
    mu sync.Mutex
    persister *btclient.Persister

    files map[TorrentMetadata]string  // map from torrent metadata paths to their local download paths
    seeding []TorrentMetadata
    shutdown chan bool
}

func StartBTClient(persister *Persister) *BTClient {
    cl := &BTClient{}
    cl.persister = persister
    cl.files = make(map[TorrentMetadata]string)
    cl.seeding = []TorrentMetadata{}
    cl.shutdown = make(chan bool)

    go cl.main()
    return cl
}

func (cl *BTClient) Kill() {
    close(cl.shutdown)
}

func (cl *BTClient) contactTracker(url string) {
    http.Get(url)
}

func (cl *BTClient) seed() {
    for {
        select {
            case _, ok := <- rf.shutdown:
                if !ok {
                    return
                }
            default:
        }
        cl.mu.Lock()
        for _, file := range cl.seeding {
            url := "" // TODO get from file
            go cl.contactTracker(url)

        }
        cl.mu.Unlock()
        time.Sleep(1 * time.Second)
    }
}

func (cl *BTClient) main() {
    go cl.seed()
    for {
        select {
            case _, ok := <- rf.shutdown:
                if !ok {
                    return
                }
            default:
        }
        time.Sleep(10 * time.Millisecond)
    }
}