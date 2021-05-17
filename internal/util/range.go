package util

import "fmt"

var errBadRange = fmt.Errorf("specified range is not good")

//FIXME: what should be semantic of range arguments ? Currently, only (0,0) args represent an infinite range;
// but what if we want [3, +inf[ ? I think we should either forbid infinite ranges
// or add a specific flag (because a [0, 0] range could be expected to be empty and is misleading)
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
