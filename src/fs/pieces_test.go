package fs

import "testing"
import "util"


func TestSplitAndCombine(t *testing.T)  {
    util.StartTest("Starting SplitAndCombine...")
    imgPath := "../main/seed/IMG_4484.cr2"
    pieceLen := 32768
    totalLen := int64(21459874)
    pieces := SplitIntoPieces(imgPath, pieceLen)
    testFilePath := "tmp.cr2"
    CombinePieces(testFilePath, pieces, totalLen)

    util.EndTest()
}
