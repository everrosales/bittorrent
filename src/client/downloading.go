package btclient

import (
	"time"
	"util"
)

func (cl *BTClient) downloadPiece(piece int) {
	cl.lock("downloader/downloadPiece")
	if _, ok := cl.blockBitmap[piece]; !ok {
		cl.blockBitmap[piece] = make([]bool, cl.numBlocks(piece), cl.numBlocks(piece))
	}
	cl.unlock("downloader/downloadPiece")
	for i := 0; i < cl.numBlocks(piece); i++ {
		cl.requestBlock(piece, i)
	}
}

// grab pieces off the needed queue and try downloading them
// re-adds pieces to queue if they weren't successfully downloaded
func (cl *BTClient) downloadPieces() {
	for {
		if cl.CheckShutdown() {
			return
		}

		piece := <-cl.neededPieces
		cl.lock("downloading/downloadPieces 1")
		if !cl.PieceBitmap[piece] {
			cl.unlock("downloading/downloadPieces 1")
			util.TPrintf("%s: trying to download piece %d\n", cl.port, piece)
			cl.downloadPiece(piece)
			cl.lock("downloading/downloadPieces 1")
		}
		cl.unlock("downloading/downloadPieces 1")

		cl.waitUntilDownloaded(piece)

		cl.lock("downloading/downloadPieces 2")
		if !cl.PieceBitmap[piece] {
			util.TPrintf("%s: piece %d was not downloaded\n", cl.port, piece)
			// piece still not downloaded, add it back to queue
			go func() {
				cl.neededPieces <- piece
			}()
		}
		cl.unlock("downloading/downloadPieces 2")
	}
}

func (cl *BTClient) waitUntilDownloaded(piece int) {
	downloaded := make(chan bool, 1)
	go func() {
		for {
			cl.lock("downloading/waitUntilDownloaded")
			done := cl.PieceBitmap[piece]
			cl.unlock("downloading/waitUntilDownloaded")
			if done {
				downloaded <- true
				return
			}
			util.Wait(10)
		}
	}()

	select {
	case <-downloaded:
	case <-time.After(time.Millisecond * 500):
	}
}
