package fs

import (
	"crypto/sha1"
	"io/ioutil"
	"math"
)

const BlockSize int = 16384

type Block []byte

type Piece struct {
	Blocks []Block
}

// get the SHA1 hash of a piece
func (piece *Piece) Hash() string {
	allBytes := []byte{}
	for _, block := range piece.Blocks {
		allBytes = append(allBytes, block...)
	}
	sha := sha1.Sum(allBytes)
	n := len(sha)
	if n != 20 {
		panic("SHA hash generation failed")
	}
	shaStr := string(sha[:n])
	return shaStr
}

// split a file into pieces of size pieceLen
func SplitIntoPieces(path string, pieceLen int) []Piece {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	numPieces := NumPieces(pieceLen, len(fileBytes))
	pieces := []Piece{}
	for i := 0; i < numPieces; i++ {
		pieces = append(pieces, getPiece(i, pieceLen, fileBytes))
	}
	return pieces
}

// combine a slice of pieces into one file, written out at path
func CombinePieces(path string, pieces []Piece, totalLen int64) {
	data := make([]byte, 0, totalLen)
	for _, piece := range pieces {
		for _, block := range piece.Blocks {
			data = append(data, block...)
		}
	}
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		panic(err)
	}
}

// returns the number of blocks for a given piece
func NumBlocksInPiece(piece int, pieceLen int, totalLen int) int {
	var actualLen int
	numPieces := NumPieces(pieceLen, totalLen)
	if piece == numPieces-1 && totalLen%pieceLen != 0 {
		// the last piece is irregular
		actualLen = totalLen % pieceLen
	} else {
		actualLen = pieceLen
	}
	return int(math.Ceil(float64(actualLen) / float64(BlockSize)))
}

// get the number of pieces given total file size and anticipated piece size
func NumPieces(pieceLen int, totalLen int) int {
	return int(math.Ceil(float64(totalLen) / float64(pieceLen)))
}

// get the truncated array from start index of at most length length
func getSubArray(start int, length int, arr []byte) []byte {
	return arr[start:min(start+length, len(arr))]
}

// get a piece from arr with index num
func getPiece(num int, pieceLen int, arr []byte) Piece {
	pieceBytes := getSubArray(num*pieceLen, pieceLen, arr)
	numBlocks := NumBlocksInPiece(num, pieceLen, len(arr))
	blocks := []Block{}
	for i := 0; i < numBlocks; i++ {
		blocks = append(blocks, getBlock(i, pieceBytes))
	}
	return Piece{blocks}
}

// get one block
func getBlock(block int, piece []byte) Block {
	start := block * BlockSize
	return Block(getSubArray(start, BlockSize, piece))
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
