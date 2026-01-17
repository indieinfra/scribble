package util

import "strings"

func FirstN[T any](s []T, n int) []T {
	if n > len(s) {
		n = len(s)
	}

	return s[:n]
}

func FirstNWords(s string, n int) []string {
	words := strings.Fields(s)
	return FirstN(words, n)
}
