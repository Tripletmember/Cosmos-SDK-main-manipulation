package types

import (
	"fmt"
)

type SecondaryKey struct {
	ID    byte
	Value []byte
}

func (s SecondaryKey) String() string {
	return fmt.Sprintf("(id=%x, value=%x)", s.ID, s.Value)
}

type Object interface {
	SecondaryKeys() []SecondaryKey
	PrimaryKey() []byte
}

type Descriptor struct {
	PrimaryKey    []byte
	SecondaryKeys map[string]SecondaryKey
}
