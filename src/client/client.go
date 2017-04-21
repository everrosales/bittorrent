package btclient
//Client

import "sync"

type BTClient struct {
    mu sync.Mutex
    persister *btclient.Persister
    shutdown chan bool
}

func StartBTClient(persister *Persister) *BTClient {
    cl := &BTClient{}
    cl.persister = persister
    cl.shutdown = make(chan bool)

    go cl.main()
    return cl
}

func (cl *BTClient) Kill() {
    close(cl.shutdown)
}

func (cl *BTClient) main() {
    for {
        time.Sleep(10 * time.Millisecond)
    }
}