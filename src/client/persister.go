package btclient

// Interface for client to save persistent state
// Modified from 6.824 raft/persister.go

import "sync"

type Persister struct {
	mu        sync.Mutex
	state []byte
}

func MakePersister() *Persister {
	return &Persister{}
}

func (ps *Persister) Copy() *Persister {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	np := MakePersister()
	np.raftstate = ps.state
	return np
}

func (ps *Persister) SaveState(data []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.state = data
}

func (ps *Persister) ReadState() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.state
}

func (ps *Persister) StateSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.state)
}
