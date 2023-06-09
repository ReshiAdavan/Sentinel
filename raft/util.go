package raft

import "log"

const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
