package main

import (
	"strconv"
)

// miscellaneous utility functions

func integerWithMinimum(str string, min int) int {
	val, err := strconv.Atoi(str)

	// fallback for invalid or nonsensical values
	if err != nil || val < min {
		val = min
	}

	return val
}

func firstElementOf(s []string) string {
	// return first element of slice, or blank string if empty
	val := ""

	if len(s) > 0 {
		val = s[0]
	}

	return val
}

func sliceContainsString(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}

func removeEntries(haystack []string, needles []string) []string {
	var res []string

	for _, hay := range haystack {
		remove := false

		for _, needle := range needles {
			if hay == needle {
				remove = true
				break
			}
		}

		if remove == true {
			continue
		}

		res = append(res, hay)
	}

	return res
}
