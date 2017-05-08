package btclient

import (
	"fs"
	"util"
)

func (cl *BTClient) lock(msg string) {
	util.LPrintf("[LOCK] %s\n", msg)
	cl.mu.Lock()
}

func (cl *BTClient) unlock(msg string) {
	util.LPrintf("[UNLOCK] %s\n", msg)
	cl.mu.Unlock()
}

func (cl *BTClient) numBlocks(piece int) int {
	return fs.NumBlocksInPiece(piece, int(cl.torrentMeta.PieceLen), cl.torrentMeta.GetLength())
}

func allTrue(arr []bool) bool {
	for _, entry := range arr {
		if !entry {
			return false
		}
	}
	return true
}
