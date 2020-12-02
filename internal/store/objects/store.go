package objects

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
)

// NewStore is Store's constructor
func NewStore(cdc codec.Marshaler, db sdk.KVStore) Store {
	return Store{
		db:  db,
		cdc: cdc,
	}
}

// Store builds an object store
type Store struct {
	db  sdk.KVStore
	cdc codec.Marshaler
}

// Create creates the object in the store
// returns an error if it already exists
// or if marshalling fails
func (s Store) Create(o types.Object) error {
	pk := o.PrimaryKey()
	if s.db.Has(pk) {
		return fmt.Errorf("%w: primary key %x", types.ErrAlreadyExists, pk)
	}
	err := s.set(pk, o)
	return err
}

// Store retrieves the object given its primary key
// fails if it does not exist, or if unmarshalling fails
// the types.Object must be a pointer
func (s Store) Read(pk []byte, o types.Object) error {
	b := s.db.Get(pk)
	// if nil we assume it was not found
	if b == nil {
		return fmt.Errorf("%w: primary key %x", types.ErrNotFound, pk)
	}
	err := s.cdc.UnmarshalBinaryLengthPrefixed(b, o)
	if err != nil {
		return err
	}
	return nil
}

// Update updates the given object, fails if it does not exist
// or if marshalling fails
func (s Store) Update(o types.Object) error {
	pk := o.PrimaryKey()
	if !s.db.Has(pk) {
		return fmt.Errorf("%w: primary key %x", types.ErrNotFound, pk)
	}
	err := s.set(pk, o)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes an object given its primary key
// fails if the object does not exist
func (s Store) Delete(primaryKey []byte) error {
	if !s.db.Has(primaryKey) {
		return fmt.Errorf("%w: primary key %x", types.ErrNotFound, primaryKey)
	}
	s.db.Delete(primaryKey)
	return nil
}

// set takes care of doing object marshalling
// and setting it in the store
func (s Store) set(key []byte, o codec.ProtoMarshaler) error {
	b, err := s.cdc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		return err
	}
	s.db.Set(key, b)
	return nil
}
