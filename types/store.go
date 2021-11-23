package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	crud "github.com/iov-one/cosmos-sdk-crud"
	"github.com/iov-one/cosmos-sdk-crud/internal/query"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/indexes"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/metadata"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/objects"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
)

// DefaultVerifyType asserts that the type is not verified when
// interacting with the store
const DefaultVerifyType = false

// ObjectsPrefix defines at which prefix of the kv store
// we are actually saving the concrete objects
const ObjectsPrefix = 0x0

// IndexesPrefix defines the prefix of the kv store
// in which we are storing indexes data
const IndexesPrefix = 0x1

// MetadataPrefix defines the prefix of the kv store
// in which we are storing objects metadata
const MetadataPrefix = 0x2

type Store struct {
	cdc codec.Codec

	verifyType bool

	objects  objects.Store
	indexes  indexes.Store
	metadata metadata.Store
}

func NewStore(cdc codec.Codec, db sdk.KVStore, pfx []byte, options ...crud.OptionFunc) Store {
	prefixedStore := prefix.NewStore(db, pfx)
	s := Store{
		cdc:        cdc,
		verifyType: DefaultVerifyType,
		objects:    objects.NewStore(cdc, prefix.NewStore(prefixedStore, []byte{ObjectsPrefix})),
		indexes:    indexes.NewStore(cdc, prefix.NewStore(prefixedStore, []byte{IndexesPrefix})),
		metadata:   metadata.NewStore(cdc, prefix.NewStore(prefixedStore, []byte{MetadataPrefix})),
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

func (s Store) Create(o crud.Object) error {
	err := s.objects.Create(o)
	if err != nil {
		return err
	}
	// create indexes
	err = s.indexes.Index(o)
	if err != nil {
		err2 := s.objects.Delete(o.PrimaryKey())
		if err2 != nil {
			panic(fmt.Errorf("state corruption unable to rollback delete after error %s: %s", err, err2))
		}
		return err
	}
	// done
	return nil
}

// Read reads the object identified by primaryKey and store it to o
// o must be an already allocated object
// Returns ErrNotFound if primaryKey identifies no object in the store
func (s Store) Read(primaryKey []byte, o crud.Object) error {
	return s.objects.Read(primaryKey, o)
}

func (s Store) Update(o crud.Object) error {
	// update indexes
	err := s.indexes.Delete(o.PrimaryKey())
	if err != nil {
		return err
	}
	err = s.indexes.Index(o)
	if err != nil {
		// state corruption, cannot rollback TODO make rollback possible
		panic(err)
	}
	err = s.objects.Update(o)
	if err != nil {
		// state corruption panic
		panic(err)
	}
	return nil
}

func (s Store) Delete(primaryKey []byte) error {
	err := s.indexes.Delete(primaryKey)
	if err != nil {
		return err
	}
	err = s.objects.Delete(primaryKey)
	if err != nil {
		// state corruption, cannot rollback. todo make rollback possible
		panic(err)
	}
	return nil
}

func (s Store) Query() crud.QueryStatement {
	return query.NewQuery(s)
}

// DoDirectQuery is used by the query package, the Query method is a more convenient way to query objects
func (s Store) DoDirectQuery(sks []crud.SecondaryKey, start, end uint64) (crud.Cursor, error) {
	var err error
	var it types.Iterator
	if len(sks) == 0 {
		it, err = s.objects.GetAllKeysWithIterator(start, end)
	} else {
		it, err = s.indexes.FilterWithIterator(sks, start, end)
	}
	if err != nil {
		return nil, err
	}
	return newFilter(it, &s), nil
}

func newFilter(it types.Iterator, store *Store) *Cursor {
	return &Cursor{
		keyIterator: it,
		store:       store,
	}
}

type Cursor struct {
	keyIterator types.Iterator
	store       *Store
}

// Next steps to the NextValue element of this cursor
func (c *Cursor) Next() {
	c.keyIterator.Next()
}

// Read reads the current element of this cursor and store it to o
// o must be an already allocated object
func (c *Cursor) Read(o crud.Object) error {
	return c.store.Read(c.currKey(), o)
}

// Delete deletes the current element of this cursor
// Delete, Read or Update should not be called on this cursor before a call to Next and will cause a ErrNotFound error
func (c *Cursor) Delete() error {
	return c.store.Delete(c.currKey())
}

// Update updates the current element of this cursor with the given object
// Delete, Read or Update should not be called on this cursor before a call to Next and may cause a ErrNotFound error
func (c *Cursor) Update(o crud.Object) error {
	return c.store.Update(o)
}

// Valid indicates if there is remaining data for this cursor
func (c *Cursor) Valid() bool {
	return c.keyIterator.Valid()
}

func (c *Cursor) currKey() []byte {
	return c.keyIterator.Get()
}
