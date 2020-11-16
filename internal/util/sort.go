package util

import (
	"bytes"
	"sort"
)

// SortByteSlice sorts a byte slice deterministically
func SortByteSlice(slice [][]byte) {
	sort.Slice(slice, func(i, j int) bool {
		return bytes.Compare(slice[i], slice[j]) < 0
	})
}

func BytesBigger(a, b []byte) bool {
	return bytes.Compare(a, b) == 1
}

func BytesSmaller(a, b []byte) bool {
	return !BytesBigger(a, b)
}

func BytesBiggerEqual(a, b []byte) bool {
	x := bytes.Compare(a, b)
	return x == 0 || x == 1
}
