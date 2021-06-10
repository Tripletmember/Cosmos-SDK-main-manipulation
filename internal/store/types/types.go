package types

import (
	crud "github.com/iov-one/cosmos-sdk-crud"
)

type Descriptor struct {
	PrimaryKey    []byte
	SecondaryKeys map[string]crud.SecondaryKey
}

type Iterator interface {
	Next()
	Valid() bool
	Get() []byte
	Collect() [][]byte
}
