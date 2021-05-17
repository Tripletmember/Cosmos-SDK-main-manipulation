package util

import "fmt"

var errBadRange = fmt.Errorf("specified range is not good")

//FIXME: what should be semantic of range arguments ? Currently, only (0,0) args represent an infinite range;
// but what if we want [3, +inf[ ? I think we should either forbid infinite ranges
// or add a specific flag (because a [0, 0] range could be expected to be empty and is misleading)
func NewRange(start, end uint64) (*Range, error) {
	if start > end {
		return nil, fmt.Errorf("%w: start bigger than end", errBadRange)
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
	// if both are zero then it's always true
	if r.start == 0 && r.end == 0 {
		return true, false
	}
	// check if we need to stop iterating
	stopIter = r.index >= r.end
	// check if we're in range
	inRange = r.index >= r.start && !stopIter
	// always bump index...
	r.index++
	// ...before returning
	return inRange, stopIter
}
