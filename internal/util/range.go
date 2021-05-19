package util

import "fmt"

var errBadRange = fmt.Errorf("specified range is not good")

// NewRange Create a new range [start, end[. Empty range are not allowed. The special value 0 for the end index
// represents a never-ending range
func NewRange(start, end uint64) (*Range, error) {
	// Empty ranges are forbidden (but when end equals 0, this is not an empty but an infinite range)
	if end != 0 && start >= end {
		return nil, fmt.Errorf("%w: empty range", errBadRange)
	}
	return &Range{
		start: start,
		end:   end,
		index: 0,
	}, nil
}

type Range struct {
	start, end uint64
	index      uint64
}

// CheckAndMoveForward Returns current range status and moves the internal iterator forward
// inRange is true if the iterator is in the range [start, end[
// stopIter is true if the iterator reached the end of the interval
func (r *Range) CheckAndMoveForward() (inRange bool, stopIter bool) {
	// check if we need to stop iterating, thus if we reached the end
	// If end == 0, then the range is infinite and should never stop iterating
	stopIter = r.end != 0 && r.index >= r.end
	// check if we're in range
	inRange = r.index >= r.start && !stopIter
	// always bump index...
	r.index++
	// ...before returning
	return inRange, stopIter
}
