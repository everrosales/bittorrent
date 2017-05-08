package btclient

import (
	"fs"
)

func (cl *BTClient) Seed(file string) {
	cl.lock("seeding/seed 1")
	pieceLen := int(cl.torrentMeta.PieceLen)
	cl.unlock("seeding/seed 1")

	pieces := fs.SplitIntoPieces(file, pieceLen)

	cl.lock("seeding/seed 2")
	copy(cl.Pieces, pieces)
	for i := range cl.PieceBitmap {
		cl.PieceBitmap[i] = true
	}
	pieceBitmap := make([]bool, len(cl.PieceBitmap))
	copy(pieceBitmap, cl.PieceBitmap)
	cl.unlock("seeding/seed 2")

	cl.persister.persistPieces(pieces, pieceBitmap)
}
