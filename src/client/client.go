package btclient
//Client

import "sync"

type BTClient struct {
    mu sync.Mutex
    shutdown chan bool
}

func StartBTClient() *BTClient {
    client := &BTClient{}
    rf.shutdown = make(chan bool)

    go client.main()
    return client
}

func (cl *BTClient) Kill() {
    close(cl.shutdown)
}

func (cl *BTClient) main() {
    for {
        time.Sleep(10 * time.Millisecond)
    }
}