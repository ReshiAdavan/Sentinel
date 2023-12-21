package raft

import "log"

// Debug is a constant used to control whether debugging messages are printed.
const Debug = 1

// DPrintf is a debugging print function. It prints messages only if Debug is greater than 0.
// It takes a format string and a variable number of arguments and uses the log package to print them.
func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...) // Print the formatted string if debugging is enabled.
	}
	return
}

// min returns the minimum of two integers.
// It is a utility function used in various places in the Raft implementation.
func min(x, y int) int {
	if x < y {
		return x // Return x if it is less than y.
	}
	return y // Otherwise, return y.
}
