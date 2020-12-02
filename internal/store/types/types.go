package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

type SecondaryKey struct {
	ID    byte
	Value []byte
}

func (s SecondaryKey) String() string {
	return fmt.Sprintf("(id=%x, value=%x)", s.ID, s.Value)
}

type Object interface {
	codec.ProtoMarshaler

	SecondaryKeys() []SecondaryKey
	PrimaryKey() []byte
}

type Descriptor struct {
	PrimaryKey    []byte
	SecondaryKeys map[string]SecondaryKey
}
