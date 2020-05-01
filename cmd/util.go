package main

import (
	"strconv"
)

// miscellaneous utility functions

func timeoutWithMinimum(str string, min int) int {
	val, err := strconv.Atoi(str)

	// fallback for invalid or nonsensical timeout values
	if err != nil || val < min {
		val = min
	}

	return val
}
