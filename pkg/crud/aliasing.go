package crud

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store"
	"github.com/iov-one/cosmos-sdk-crud/pkg/crud/types"
)

type Store types.Store
type PrimaryKey types.PrimaryKey
type SecondaryKey types.SecondaryKey

// NewStore returns a new CRUD key value store
func NewStore(ctx sdk.Context, key sdk.StoreKey, cdc *codec.Codec, uniquePrefix []byte) types.Store {
	return store.NewStore(ctx, key, cdc, uniquePrefix)
}

// NewPrimaryKey aliases types.NewPrimaryKey
func NewPrimaryKey(key []byte) types.PrimaryKey {
	return types.NewPrimaryKey(key)
}

// NewPrimaryKeyFromString aliases types.NewPrimaryKeyFromString
func NewPrimaryKeyFromString(str string) types.PrimaryKey {
	return types.NewPrimaryKeyFromString(str)
}

// NewSecondaryKey aliases types.NewSecondaryKey
func NewSecondaryKey(storePrefix byte, key []byte) types.SecondaryKey {
	return types.NewSecondaryKey(storePrefix, key)
}

// NewSecondaryKeyFromBytes aliases types.NewSecondaryKeyFromBytes
func NewSecondaryKeyFromBytes(b []byte) types.SecondaryKey {
	return types.NewSecondaryKeyFromBytes(b)
}
