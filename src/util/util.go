package util

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type color string
type ansiControl string
type debugLevel int

// global var set in main for debug level
var Debug debugLevel

const (
	Lock  debugLevel = 1
	Trace debugLevel = 2
	Info  debugLevel = 3
	None  debugLevel = 4
)

const (
	Default   color = ""
	Underline color = "\033[4m"
	Red       color = "\033[31m"
	Green     color = "\033[32m"
	Yellow    color = "\033[33m"
	Blue      color = "\033[34m"
	Purple    color = "\033[35m"
	Cyan      color = "\033[36m"
	Reset     color = "\033[0m"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func StartTest(desc string) {
	if Debug != None {
		desc = desc + "\n"
	}
	ColorPrintf(Underline, desc)
}

func EndTest() {
	ColorPrintf(Green, "  pass\n")
}

// default printing
func Printf(format string, a ...interface{}) (n int, err error) {
	ColorPrintf(Default, format, a...)
	return
}

func Printfln(format string, a ...interface{}) (n int, err error) {
	ColorPrintf(Default, format+"\n", a...)
	return
}

// error logging
func EPrintf(format string, a ...interface{}) (n int, err error) {
	ColorPrintf(Red, "[ERROR] "+format, a...)
	return
}

// warning logging
func WPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Info {
		ColorPrintf(Yellow, "[WARNING] "+format, a...)
	}
	return
}

// info logging
func IPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Info {
		ColorPrintf(Blue, format, a...)
	}
	return
}

// trace logging
func TPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Trace {
		format = strings.Replace(format, "\n", "\n    ", strings.Count(format, "\n")-1)
		ColorPrintf(Default, "    "+format, a...)
	}
	return
}

//lock logging
func LPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Lock {
		ColorPrintf(Purple, format, a...)
	}
	return
}

// generic printing with color
func ColorPrintf(c color, format string, a ...interface{}) (n int, err error) {
	str := string(c) + format + string(Reset)
	fmt.Printf(str, a...)
	// fmt.Flush()
	os.Stdout.Sync()
	return
}

func Wait(milliseconds int) {
	<-time.After(time.Millisecond * time.Duration(milliseconds))
}

func ByteArrayEquals(first []byte, second []byte) bool {
	if first == nil && second == nil {
		return true
	}
	if first == nil || second == nil {
		return false
	}
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func BoolArrayEquals(first []bool, second []bool) bool {
	if first == nil && second == nil {
		return true
	}
	if first == nil || second == nil {
		return false
	}
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func SplitEveryN(str string, n int) []string {
	slices := []string{}
	lastIndex := 0
	for i := range str {
		if i == len(str)-1 {
			slices = append(slices, str[lastIndex:])
		} else if i-lastIndex == 20 {
			slices = append(slices, str[lastIndex:i])
			lastIndex = i
		}
	}
	return slices
}

func GenerateRandStr(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func BoolsToBytes(data []bool) []byte {
	if len(data)%8 != 0 {
		// we need to pad the message with zeros
		padBuf := make([]bool, 8-(len(data)%8))
		// the extra pad needs to be false (0-padded)
		data = append(data, padBuf...)
	}

	if len(data)%8 != 0 {
		EPrintf("boolsToBytes: wtf")
	}

	output := make([]byte, len(data)/8)
	for i := 0; i < len(data); i++ {
		val := byte(0)
		if data[i] {
			val = byte(0x80)
		}
		output[i/8] = output[i/8] | (val >> uint(i%8))
	}
	return output
}

func BytesToBools(data []byte) []bool {
	output := make([]bool, len(data)*8)
	for i := 0; i < len(data)*8; i++ {
		mask := (byte(0x80) >> uint(i%8))
		if (data[i/8] & mask) > 0 {
			output[i] = true
		}
		// We dont have to set the false bits because thats done for us by the make operation
	}
	return output
}

func BitfieldToString(data []bool, width int) (string, int) {
	sum := 0.0
	piecePercent := float64(width) / float64(len(data))
	result := "["

	for i := 0; i < len(data); i++ {
		if data[i] {
			sum += piecePercent
		}
	}

	for i := 0; i < int(sum); i++ {
		result += "#"
	}
	for i := int(sum); i < width; i++ {
		result += "-"
	}
	percent := sum * 100 / float64(width)
	result += fmt.Sprintf("] %.2f percent", percent)

	return result, 0 // number of newlines
}

func MoveCursorUp(numlines int) {
	Printf("\033[1000D")
	for i := 0; i < numlines; i++ {
		Printf("\033[1A")
	}
}

func ZeroCursor() {
	// Printf("\033[0G")
	Printf("\033[0;0H")
}

func MoveCursorDown(numlines int) {
	// Printf("\033[1000D")
	for i := 0; i < numlines; i++ {
		Printf("\033[1B")
	}
}

func ClearScreen() {
	Printf("\033[0J")
}
