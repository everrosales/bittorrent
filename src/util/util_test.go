package util

import "testing"

//            d                       e                        a                         d
//            1     1     0      1    1     1     1     0      1     0      1     0      1     1     0      1
//            1     0     1      1    1     1     1     0      1     1     1     0      1     1     1     1
var deadbeefBools []bool = []bool{true, true, false, true, true, true, true, false, true, false, true, false, true, true, false, true,
             true, false, true, true, true, true, true, false, true, true, true, false, true, true, true, true}
var deadbeefBytes []byte = []byte{0xde, 0xad, 0xbe, 0xef}
//            d                        e                        a                         0
//            1     1     0      1     1     1     1     0      1     0      1     0      0
var halfdeadBools []bool =  []bool{true, true, false, true, true, true, true, false, true, false, true, false, false }
var halfdeadBytes []byte= []byte{0xde, 0xa0}

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
