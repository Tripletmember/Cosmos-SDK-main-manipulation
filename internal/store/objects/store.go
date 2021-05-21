package objects

import (
	"fmt"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/iterator"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/util"
)

// NewStore is Store's constructor
func NewStore(cdc codec.Marshaler, db sdk.KVStore) Store {
	ctr := Counter{}
	return Store{
		db:      db,
		cdc:     cdc,
		objects: &ctr,
	}
}

type Counter struct {
	count uint64
}

// Store builds an object store
type Store struct {
	db  sdk.KVStore
	cdc codec.Marshaler
	// This has to be a reference in order to persist between calls
	// Otherwise, if we want to use a simple uint64,
	// we must switch to pointer receiver and pointer storage in struct (or use it through an interface)
	objects *Counter
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
	s.objects.count++
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
	// TODO: Could an user alter the primary key of an object ? If this is the case,
	// update semantics are quite strange
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
	s.objects.count--
	return nil
}

// GetAllKeysWithIterator returns an iterator yielding the primary key of all the objects present in the store
// in the interval [start, end[ and in ascending order.
func (s Store) GetAllKeysWithIterator(start uint64, end uint64) (types.Iterator, error) {
	// We could use append but it has to reallocate each time its capacity is reached
	// Tracking the number of objects on the store is more efficient

	// If the start offset is superior to the number of element, we can skip directly
	if start >= s.objects.count {
		return iterator.NilIterator{}, nil
	}

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
