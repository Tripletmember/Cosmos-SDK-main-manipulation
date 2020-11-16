package types

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when an index or an object are not found
var ErrNotFound = errors.New("crud: not found")

// ErrAlreadyExists is returned when an index or an object already exist
var ErrAlreadyExists = errors.New("crud: already exists")

// ErrBadArgument is returned when the provided arguments are invalid
var ErrBadArgument = errors.New("crud: bad argument")

// ErrInternal is returned when the store detects internal error which
// might be related to possible state corruption
var ErrInternal = errors.New("crud: internal error")

// ErrCursorConsumed is returned in case the cursor used for primary key filtering
// is not valid anymore because it was consumed
var ErrCursorConsumed = fmt.Errorf("%w: cursor consumed", ErrBadArgument)
