package metadata

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Store struct {
	cdc codec.Codec
	db  sdk.KVStore
}

func NewStore(cdc codec.Codec, db sdk.KVStore) Store {
	return Store{
		cdc: cdc,
		db:  db,
	}
}
