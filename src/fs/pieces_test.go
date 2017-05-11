package fs

import (
	"testing"
	"util"
)

func init() {
	util.Debug = util.None
}

func TestSplitAndCombineFile(t *testing.T) {
	util.StartTest("Testing splitting and recombining a file...")
	imgPath := "../main/seed/IMG_4484.cr2"
	pieceLen := 32768
	totalLen := int64(21459874)
	pieces := SplitIntoPieces(imgPath, pieceLen)
	testFilePath := "tmp.cr2"
	CombinePieces(testFilePath, pieces, totalLen)

	same, err := util.CompareFiles("tmp.cr2", "../main/seed/IMG_4484.cr2")

	if err != nil || !same {
		t.Fatalf("Split and recombined files don't match")
	}

	util.EndTest()
}
