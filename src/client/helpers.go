package btclient

import (
	"btnet"
	"fmt"
	"fs"
	"math/rand"
	"util"
)

func (cl *BTClient) lock(msg string) {
	util.LPrintf("[LOCK] %s: %s\n", cl.port, msg)
	cl.mu.Lock()
}

func (cl *BTClient) unlock(msg string) {
	util.LPrintf("[UNLOCK] %s: %s\n", cl.port, msg)
	cl.mu.Unlock()
}

func (cl *BTClient) numBlocks(piece int) int {
	return fs.NumBlocksInPiece(piece, int(cl.torrentMeta.PieceLen), cl.torrentMeta.GetLength())
}

func (cl *BTClient) getRandomPeerOrder() []*btnet.Peer {
	peerList := make([]*btnet.Peer, len(cl.peers))
	order := rand.Perm(len(peerList))
	i := 0
	// create a list of peers in random order
	for addr := range cl.peers {
		peerList[order[i]] = cl.peers[addr]
		i += 1
	}
	return peerList
}

func (cl *BTClient) atomicAddUpdate(addr string, piece int, verb string) {
	cl.lock("adding received update")
	defer cl.unlock("adding received update")
	cl.updates = append(cl.updates[1:], fmt.Sprintf("%s block in piece %d - %s", verb, piece, addr))
}

func (cl *BTClient) AtomicGetBitmap() []bool {
	cl.lock("getting bitmap element")
	defer cl.unlock("getting bitmap element")
	result := make([]bool, len(cl.PieceBitmap))
	copy(result, cl.PieceBitmap)
	return result
}

func (cl *BTClient) atomicGetBitmapElement(index int) bool {
	cl.lock("getting bitmap element")
	defer cl.unlock("getting bitmap element")
	return cl.PieceBitmap[index]
}

func (cl *BTClient) atomicGetPeer(addr string) (*btnet.Peer, bool) {
	cl.lock("client/atomicGetPeer")
	p, ok := cl.peers[addr]
	cl.unlock("client/atomicGetPeer")
	return p, ok
}

func (cl *BTClient) atomicSetPeer(addr string, peer *btnet.Peer) {
	cl.lock("client/atomicSetPeer")
	cl.peers[addr] = peer
	cl.unlock("client/atomicSetPeer")
	return
}

func (cl *BTClient) atomicDeletePeer(addr string) {
	cl.lock("client/atomicDeletePeer")
	util.WPrintf("%s: keepalive timeout exceeded for %s\n", cl.port, addr)
	delete(cl.peers, addr)
	cl.unlock("client/atomicDeletePeer")
}

func (cl *BTClient) atomicGetPeerAddrs() []string {
	cl.lock("client/atomicGetPeerAddrs")
	result := []string{}
	for addr := range cl.peers {
		result = append(result, addr)
	}
	cl.unlock("client/atomicGetPeerAddrs")
	return result
}

func (cl *BTClient) atomicGetNumPeers() int {
	cl.lock("client/atomicGetNumPeers")
	num := len(cl.peers)
	cl.unlock("client/atomicGetNumPeers")
	return num
}
