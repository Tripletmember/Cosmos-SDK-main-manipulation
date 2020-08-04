package store

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/filter"
	"github.com/iov-one/cosmos-sdk-crud/pkg/crud/types"
)

var indexListPrefix = []byte{0x2}
var indexPrefix = []byte{0x01}
var objectPrefix = []byte{0x00}

type objects interface {
	create(o types.Object)
	read(key types.PrimaryKey, o types.Object) bool
	delete(key types.PrimaryKey)
	update(o types.Object)
	iterate(do func(pk types.PrimaryKey) bool)
}

type indexes interface {
	create(o types.Object)
	delete(pk types.PrimaryKey)
	iterate(sk types.SecondaryKey, do func(pk types.PrimaryKey) bool)
}

// Store defines a crud object store
// the store creates two sub-stores
// using prefixing, one is used to store objects
// the other one is used to store the indexes of
// the object.
type Store struct {
	objects objects
	indexes indexes
	raw     sdk.KVStore
}

// NewStore generates a new crud.Store given a context, a store key, the codec and a unique prefix
// that can be specified as nil if not required, the prefix generally serves the purpose of splitting
// a store into different stores in case different objects have to coexist in the same store.
func NewStore(ctx sdk.Context, key sdk.StoreKey, cdc *codec.Codec, uniquePrefix []byte) Store {
	store := ctx.KVStore(key)
	if len(uniquePrefix) != 0 {
		store = prefix.NewStore(store, uniquePrefix)
	}
	return Store{
		indexes: newIndexes(cdc, store),
		objects: newObjectsStore(cdc, store),
		raw:     store,
	}
}

func (s Store) Filter(fltr types.Object) types.Filter {
	// if the primary key is specified then return that
	var primaryKeys []types.PrimaryKey
	pk := fltr.PrimaryKey()
	// if primary key exists then just use that
	if pk != nil && len(pk.Key()) != 0 {
		primaryKeys = append(primaryKeys, pk)
		return filter.NewFiltered(primaryKeys, s)
	}
	primaryKeys = append(primaryKeys, s.getPrimaryKeys(fltr.SecondaryKeys())...)
	return filter.NewFiltered(primaryKeys, s)
}

// Create creates a new object in the object store and writes its indexes
func (s Store) Create(o types.Object) {
	primaryKey := o.PrimaryKey()
	// TODO this in the future needs to be autogenerated to decouple
	// the need for a primary key for objects that do not need or have it
	// and rely on indexes for filtering of different objects
	if len(primaryKey.Key()) == 0 {
		panic("empty primary key provided")
	}
	// create object
	s.objects.create(o)
	// generate indexes
	s.indexes.create(o)
}

// Read reads in the object store and returns false if the object is not found
// if it is found then the binary is unmarshalled into the Object.
// CONTRACT: Object must be a pointer for the unmarshalling to take effect.
func (s Store) Read(key types.PrimaryKey, o types.Object) (ok bool) {
	return s.objects.read(key, o)
}

func (s Store) IterateKeys(do func(pk types.PrimaryKey) bool) {
	s.objects.iterate(do)
}

// Update updates the given Object in the objects store
// after clearing the indexes and reapplying them based on the
// new update.
// To achieve so a zeroed copy of Object is created which is used to
// unmarshal the old object contents which is necessary for the un-indexing.
func (s Store) Update(newObject types.Object) {
	pk := newObject.PrimaryKey()
	// remove old indexes
	s.indexes.delete(pk)
	// set new object
	s.objects.update(newObject)
	// index new object
	s.indexes.create(newObject)
}

// Delete deletes an object from the object store after
// clearing its indexes, the object is required to provide
// a kind to clone in order to remove indexes related
func (s Store) Delete(pk types.PrimaryKey) {
	// remove indexes
	s.indexes.delete(pk)
	// remove key
	s.objects.delete(pk)
}

func (s Store) getPrimaryKeys(filters []types.SecondaryKey) []types.PrimaryKey {
	sets := make([]set, 0, len(filters))
	for _, fltr := range filters {
		set := make(keySet)
		s.indexes.iterate(fltr, func(key types.PrimaryKey) bool {
			set.Insert(key)
			return true
		})
		sets = append(sets, set)
	}
	return primaryKeysFromSets(sets)
}
