package btclient

import (
	"util"
)

func (cl *BTClient) downloadPiece(piece int) {
	util.TPrintf("%s: need to download piece %d\n", cl.port, piece)
	if _, ok := cl.blockBitmap[piece]; !ok {
		cl.blockBitmap[piece] = make([]bool, cl.numBlocks(piece), cl.numBlocks(piece))
	}
	for i := 0; i < cl.numBlocks(piece); i++ {
		cl.requestBlock(piece, i)
	}
}

func (cl *BTClient) downloadPieces() {
	for {
		if cl.CheckShutdown() {
			return
		}
		piece := <-cl.neededPieces
		cl.lock("downloading/downloadPieces 1")
		if !cl.PieceBitmap[piece] {
			cl.unlock("downloading/downloadPieces 1")
			cl.downloadPiece(piece)
			cl.lock("downloading/downloadPieces 1")
		}
		cl.unlock("downloading/downloadPieces 1")

		util.Wait(200)

		cl.lock("downloading/downloadPieces 2")
		if !cl.PieceBitmap[piece] {
			util.TPrintf("%s: piece was not downloaded\n", cl.port)
			// piece still not downloaded, add it back to queue
			go func() {
				cl.neededPieces <- piece
			}()
		}
		cl.unlock("downloading/downloadPieces 2")
	}
}
