package util

import (
	"crypto/sha1"
	"testing"
)

//            d                       e                        a                         d
//            1     1     0      1    1     1     1     0      1     0      1     0      1     1     0      1
//            1     0     1      1    1     1     1     0      1     1     1     0      1     1     1     1
var deadbeefBools []bool = []bool{true, true, false, true, true, true, true, false, true, false, true, false, true, true, false, true,
	true, false, true, true, true, true, true, false, true, true, true, false, true, true, true, true}
var deadbeefBytes []byte = []byte{0xde, 0xad, 0xbe, 0xef}

//            d                        e                        a                         0
//            1     1     0      1     1     1     1     0      1     0      1     0      0
var halfdeadBools []bool = []bool{true, true, false, true, true, true, true, false, true, false, true, false, false}
var halfdeadBytes []byte = []byte{0xde, 0xa0}

func init() {
	Debug = None
}

func TestBoolsToBytes(t *testing.T) {
	StartTest("Testing BoolsToBytes ...")

	actual := BoolsToBytes(deadbeefBools)
	if !ByteArrayEquals(actual, deadbeefBytes) {
		t.Fatalf("actual != expected: no deadbeef")
	}
	// Test padding
	actual = BoolsToBytes(halfdeadBools)
	if !ByteArrayEquals(actual, halfdeadBytes) {
		t.Fatalf("actual != expected: not halfdead")
	}
	EndTest()
}

func TestBytesToBools(t *testing.T) {
	StartTest("Testing BytesToBools...")
	actual := BytesToBools(deadbeefBytes)
	if !BoolArrayEquals(actual, deadbeefBools) {
		t.Fatalf("actual != expected: no deadbeef")
	}
	EndTest()
}

func TestBitfieldToString(t *testing.T) {
	StartTest("Testing BitfieldToString...")
	bitfield := []bool{false, true, false,
		true, false, true,
		false, true, true}
	output, numLines := BitfieldToString(bitfield, 3)
	if output[:5] != "[#--]" {
		t.Fatalf("Progress bar not as expected")
	}
	if numLines != 0 {
		t.Fatalf("numLines is wrong")
	}
	EndTest()
}

func TestMoveCursorUp(t *testing.T) {
	StartTest("Testing moving cursor... (should print pass on next line)")
	Printf("\nasdf\n")
	MoveCursorUp(1)
	EndTest()
}

func TestSplitAndCombineHashes(t *testing.T) {
	StartTest("Testing splitting and recombining hashes...")
	str1 := "blahblaegre"
	str2 := "bwjgweij"
	str3 := "poop"
	shaStr1 := sha1.Sum([]byte(str1))
	if len(shaStr1) != 20 {
		t.Fatalf("SHA is not length 20")
	}
	shaStr2 := sha1.Sum([]byte(str2))
	if len(shaStr2) != 20 {
		t.Fatalf("SHA is not length 20")
	}
	shaStr3 := sha1.Sum([]byte(str3))
	if len(shaStr3) != 20 {
		t.Fatalf("SHA is not length 20")
	}
	//TPrintf("First SHA %s, second SHA %s, third SHA %s\n", shaStr1, shaStr2, shaStr3)
	newStr := string(shaStr1[:20]) + string(shaStr2[:20]) + string(shaStr3[:20]) + "aaaaa"

	if len(newStr) != 65 {
		t.Fatalf("Concatenated string length is not 65")
	}

	pieces := SplitEveryN(newStr, 20)
	TPrintf("%d %d %d %d\n", len(pieces[0]), len(pieces[1]), len(pieces[2]), len(pieces[3]))

	if len(pieces) != 4 {
		t.Fatalf("Didn't get the right number of pieces")
	}
	if len(pieces[0]) != 20 {
		t.Fatalf("First piece is not right length: %d != 20", len(pieces[0]))
	}
	if len(pieces[1]) != 20 {
		t.Fatalf("Second piece is not right length: %d != 20", len(pieces[1]))
	}
	if len(pieces[2]) != 20 {
		t.Fatalf("Third piece is not right length: %d != 20", len(pieces[2]))
	}
	if len(pieces[3]) != 5 {
		t.Fatalf("Last piece is not right length: %d != 5", len(pieces[3]))
	}
	if pieces[0] != string(shaStr1[:20]) {
		t.Fatalf("String 1 didn't match: %s != %s", pieces[0], string(shaStr1[:20]))
	}
	if pieces[1] != string(shaStr2[:20]) {
		t.Fatalf("String 2 didn't match: %s != %s", pieces[1], string(shaStr2[:20]))
	}
	if pieces[2] != string(shaStr3[:20]) {
		t.Fatalf("String 3 didn't match: %s != %s", pieces[2], string(shaStr3[:20]))
	}
	if pieces[3] != "aaaaa" {
		t.Fatalf("String 4 didn't match: %s != aaaaa", pieces[3])
	}
	EndTest()
}
