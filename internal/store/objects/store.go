package objects

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crud "github.com/iov-one/cosmos-sdk-crud"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/iterator"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/util"
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
func (s Store) Create(o crud.Object) error {
	pk := o.PrimaryKey()
	if s.db.Has(pk) {
		return fmt.Errorf("%w: primary key %x", crud.ErrAlreadyExists, pk)
	}
	err := s.set(pk, o)
	return err
}

// Store retrieves the object given its primary key
// fails if it does not exist, or if unmarshalling fails
// the crud.Object must be a pointer
func (s Store) Read(pk []byte, o crud.Object) error {

	b := s.db.Get(pk)
	// if nil we assume it was not found
	if b == nil {
		return fmt.Errorf("%w: primary key %x", crud.ErrNotFound, pk)
	}
	err := s.cdc.UnmarshalBinaryLengthPrefixed(b, o)
	if err != nil {
		return err
	}
	return nil
}

// Update updates the given object, fails if it does not exist
// or if marshalling fails
func (s Store) Update(o crud.Object) error {
	// TODO: Could an user alter the primary key of an object ? If this is the case,
	// update semantics are quite strange
	pk := o.PrimaryKey()
	if !s.db.Has(pk) {
		return fmt.Errorf("%w: primary key %x", crud.ErrNotFound, pk)
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
		return fmt.Errorf("%w: primary key %x", crud.ErrNotFound, primaryKey)
	}
	s.db.Delete(primaryKey)
	return nil
}

// GetAllKeysWithIterator returns an iterator yielding the primary key of all the objects present in the store
// in the interval [start, end[ and in ascending order.
func (s Store) GetAllKeysWithIterator(start uint64, end uint64) (types.Iterator, error) {
	// We could use append but it has to reallocate each time its capacity is reached
	// Tracking the number of objects on the store is more efficient

	// The start and end arguments of Iterator() are not indexes but byte array boundaries
	it := s.db.Iterator(nil, nil)

	rng, err := util.NewRange(start, end)
	if err != nil {
		return iterator.NilIterator{}, err
	}

	resultIterator := iterator.NewKeyIterator(func() ([]byte, bool) {
		for {
			noMoreValues := !it.Valid()
			inRange, stopIter := rng.CheckAndMoveForward()
			if stopIter || noMoreValues {
				it.Close()
				return nil, false
			}

			pk := it.Key()
			it.Next()
			if inRange {
				return pk, true
			}
		}
	})

	return resultIterator, nil
}

// GetAllKeys returns a slice containing the primary key of all the objects present in the store
// in the interval [start, end[ and in ascending order.
func (s Store) GetAllKeys(start, end uint64) ([][]byte, error) {
	it, err := s.GetAllKeysWithIterator(start, end)
	if err != nil {
		return nil, err
	}
	return it.Collect(), err
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
