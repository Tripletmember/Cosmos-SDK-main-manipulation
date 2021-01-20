package util

import "fmt"

var errBadRange = fmt.Errorf("specified range is not good")

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
	inRange = r.index >= r.start
	// always bump index...
	r.index++
	// ...before returning
	return inRange, stopIter
}
