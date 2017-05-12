package btclient

import (
	"fs"
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

func (cl *BTClient) atomicGetBitmapElement(index int) bool {
	cl.lock("getting bitmap element")
	defer cl.unlock("getting bitmap element")
	return cl.PieceBitmap[index]
}

func allTrue(arr []bool) bool {
	for _, entry := range arr {
		if !entry {
			return false
		}
	}
	return true
}
