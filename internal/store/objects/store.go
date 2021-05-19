package objects

import (
	"fmt"

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

// GetAllKeys returns a slice containing the primary key of all the objects present in the store
// in the interval [start, end[ and in ascending order.
// Returns an error if the iterator cannot be closed
func (s Store) GetAllKeys(start, end uint64) ([][]byte, error) {
	// We could use append but it has to reallocate each time its capacity is reached
	// Tracking the number of objects on the store is more efficient

	// If the start offset is superior to the number of element, we can skip directly
	if start >= s.objects.count {
		return make([][]byte, 0), nil
	}

	// The maximum size of the result set is the number of objects minus the start offset
	var size = s.objects.count - start
	// If the range is finite, the actual size is the queried length, with a maximum of the previously computed length
	if end != 0 {
		size = util.Uint64Min(end-start, size)
	}

	keys := make([][]byte, size)
	// The start and end arguments of Iterator() are not indexes but byte array boundaries
	it := s.db.Iterator(nil, nil)
	defer it.Close()

	rng, err := util.NewRange(start, end)
	if err != nil {
		return make([][]byte, 0), err
	}

	var stopIter, inRange bool
	for i := uint64(0); it.Valid() && !stopIter; it.Next() {
		inRange, stopIter = rng.CheckAndMoveForward()
		//TODO: this could be changed with juste inRange when merged with the new Range version
		if inRange && !stopIter {
			// This should never happen and is an internal error
			if i >= size {
				return make([][]byte, 0), types.ErrInternal
			}

			keys[i] = it.Key()
			i++
		}
	}
	return keys, nil
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
