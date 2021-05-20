package util

import (
	"bytes"
	"sort"
)

// SortByteSlice sorts a byte slice deterministically
func SortByteSlice(slice [][]byte) {
	// We could use SliceStable in order to have a deterministic order out of this function, but as only the
	// data is important and not the actual slices objects order, Slice is sufficient (and could be more efficient)
	sort.Slice(slice, func(i, j int) bool {
		return BytesSmaller(slice[i], slice[j])
	})
}

func BytesSmaller(a, b []byte) bool {
	return bytes.Compare(a, b) < 0
}

func BytesBiggerEqual(a, b []byte) (isGreater, isEqual bool) {
	comp := bytes.Compare(a, b)
	isGreater = comp > 0
	isEqual = comp == 0
	return
}
