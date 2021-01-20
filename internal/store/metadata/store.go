package metadata

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Store struct {
	cdc codec.Marshaler
	db  sdk.KVStore
}

func NewStore(cdc codec.Marshaler, db sdk.KVStore) Store {
	return Store{
		cdc: cdc,
		db:  db,
	}
}
