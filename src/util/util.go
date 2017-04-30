package util

import (
	"bytes"
	"fmt"
)

type color string
type debugLevel int

// global var set in main for debug level
var Debug debugLevel

const (
	Trace debugLevel = 1
	Info  debugLevel = 2
	None  debugLevel = 3
)

const (
	Default color = ""
	Red     color = "\033[31m"
	Green   color = "\033[32m"
	Yellow  color = "\033[33m"
	Blue    color = "\033[34m"
	Purple  color = "\033[35m"
	Cyan    color = "\033[36m"
	Reset   color = "\033[39m"
)

// error logging
func EPrintf(format string, a ...interface{}) (n int, err error) {
	DPrintf(Red, "[ERROR] "+format, a...)
	return
}

// info logging
func IPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Info {
		DPrintf(Blue, format, a...)
	}
	return
}

// trace logging
func TPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Trace {
		DPrintf(Default, format, a...)
	}
	return
}

// generic printing with color
func DPrintf(c color, format string, a ...interface{}) (n int, err error) {
	str := string(c) + format + string(Reset)
	fmt.Printf(str, a...)
	return
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

// from http://stackoverflow.com/questions/25686109/split-string-by-length-in-golang
func SplitEveryN(s string, n int) []string {
	sub := ""
	subs := []string{}

	runes := bytes.Runes([]byte(s))
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i+1)%n == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}

	return subs
}

func TestStartPrintf(format string, a ...interface{}) {
  str := "----------------\n" + format
  fmt.Printf(str, a...)
  return
}

func TestFinishPrintf(format string, a ...interface{}) {
  str := format + "----------------\n"
  fmt.Printf(str, a...)
  return
}
