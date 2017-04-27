package util

import (
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
