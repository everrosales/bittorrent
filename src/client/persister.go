package btclient

// Interface for client to save persistent state
// Modified from 6.824 raft/persister.go

import "sync"
import "io/ioutil"

type Persister struct {
	mu        sync.Mutex
	state []byte
	Path string
}

func MakePersister(path string) *Persister {
	return &Persister{Path: path}
}

func (ps *Persister) Copy() *Persister {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	np := MakePersister(ps.Path)
	np.state = ps.state
	return np
}

func (ps *Persister) SaveState(data []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ioutil.WriteFile(ps.Path, data, 0644)
	ps.state = data
}

func (ps *Persister) ReadState() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	data, err := ioutil.ReadFile(ps.Path)
	if err != nil {
		ps.state = nil
	} else {
		ps.state = data
	}
	return ps.state
}

func (ps *Persister) StateSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.state)
}
