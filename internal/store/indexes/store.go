package indexes

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/util"
)

// experimentalFiltering defines if to use the experimental filtering function or not
const experimentalFiltering = true

// indexesPrefix is the prefix used to save index to primary key data
const indexesPrefix = 0x0

// primaryKeysToIndexPrefix is the prefix used to map primary keys to indexes
const primaryKeysToIndexPrefix = 0x1

// Store defines the index store, it store mappings from indexes
// to primary keys, and mappings of primary keys to their respective
// index values
type Store struct {
	// codec serves the purpose of encoding and decoding objects
	cdc codec.Marshaler // TODO get rid of the codec?
	// indexes maps secondary keys to their primary keys
	// index and primary keys are stored using the following pattern
	// the index key is composed as
	// <ID><Length of index key value in little endian><IndexKeyValue>
	// where ID defines the unique identifier of the index (example, Location with index id which equals 0x0)
	// Length is the length of the index key value
	// IndexKeyValue is the value of the index we're mapping primary keys to, for example (Location) Italy
	// So when we're storing objects which we want to index by their Location the following KVStore forms
	// considering Italy as location
	// we get the following prefixed store:
	// <ID=0x0><LengthOfIndexValue=5><italy>
	// and we start saving primary keys as keys in the store, with value []byte{}
	// so, in the store, what we get is the following
	// <ID=0x0><LengthOfIndexValue=5><italy> PrimaryKey_A
	// <ID=0x0><LengthOfIndexValue=5><italy> PrimaryKey_B
	// so if we want to get all the objects which have Italy as index value
	// we just prefix the store using the index key value = italy
	// and we automatically get access to all the primary keys required
	indexes sdk.KVStore
	// primaryKeysIndexes store the encoded secondary keys values of an object
	// using its primary key as key in the store. This allows us to quickly update
	// or get rid of indexes while updating or deleting an object from the store.
	primaryKeysIndexes sdk.KVStore
}

// NewStore builds the required prefixed stores used by the index store
// one which contains the maps towards index keys to primary keys
// and the other which maps a primary key to its respective index key
// for a more straight forward delete of indexes which does not require
// acquiring the objects current state when indexes are updated.
func NewStore(cdc codec.Marshaler, db sdk.KVStore) Store {
	return Store{
		cdc:                cdc,
		indexes:            prefix.NewStore(db, []byte{indexesPrefix}),
		primaryKeysIndexes: prefix.NewStore(db, []byte{primaryKeysToIndexPrefix}),
	}
}

// Index creates, given a types.Object, it's index value to primary key
// pointers, and also the primary keys to indexes list
func (s Store) Index(o types.Object) error {
	primaryKey := o.PrimaryKey()                   // gets the object's primary key
	secondaryKeys := o.SecondaryKeys()             // gets the object's secondary keys
	keysList := make([][]byte, len(secondaryKeys)) // create the slice for computed index keys
	// iterate over secondary keys
	for i, secondaryKey := range secondaryKeys {
		// make the secondary key point to this object
		computedKey, err := s.mapKey(secondaryKey, primaryKey)
		if err != nil {
			return err
		}
		// add the computed key obtained after registering it to the keys lsit
		keysList[i] = computedKey
	}
	// save indexes list
	err := s.saveIndexList(primaryKey, keysList)
	if err != nil {
		// try to rollback what was done so far
		err2 := s.unmapRawKeys(primaryKey, keysList)
		// if it does not work panic, we got state corruption.
		if err2 != nil {
			panic(fmt.Errorf("state corruption unable to rollback index delete after error %s: %s", err, err2))
		}
	}
	return nil
}

// Delete retrieves the list of indexes which map to the given primary key
// and gets rid of them, so in future queries using the indexes the object
// is not retrieved anymore.
func (s Store) Delete(primaryKey []byte) error {
	secondaryKeys, err := s.getIndexList(primaryKey)
	err = s.unmapRawKeys(primaryKey, secondaryKeys)
	if err != nil {
		// state corruption, as we might have deleted a subset of keys
		// but not all of them... sad but true
		panic(err)
	}
	// clear index list
	err = s.deleteIndexList(primaryKey)
	if err != nil {
		return err
	}
	return nil
}

// QueryAll will return all the primary keys contained in an index, be careful
// as it will load all primary keys in memory, generally speaking Query is suggested
// for wide index queries.
func (s Store) QueryAll(sk types.SecondaryKey) (primaryKeys [][]byte, err error) {
	err = s.Query(sk, 0, 0, func(primaryKey []byte) (stop bool) {
		primaryKeys = append(primaryKeys, primaryKey)
		return false
	})
	if err != nil {
		return nil, err
	}
	return
}

// Query queries the index given a secondary key start and end defines the start and end of the
// exact part of the index which we want to query, to query the whole index domain just put start and
// end as 0 the primary keys found will be passed to the 'do' function, if 'do' returns false, the
// iteration is stopped.
func (s Store) Query(sk types.SecondaryKey, start, end uint64, do func(primaryKey []byte) (stop bool)) error {
	return s.getPrimaryKeysFromIndex(sk, start, end, do)
}

// getPrimaryKeysFromIndex gets all the primary keys from the given start-end range
// start and end are inclusive, error is returned only in case the provided secondary key is invalid
func (s Store) getPrimaryKeysFromIndex(sk types.SecondaryKey, start uint64, end uint64, do func(primaryKey []byte) (stop bool)) error {
	store, _, err := s.kvStore(sk)
	if err != nil {
		return err
	}
	iter := store.Iterator(nil, nil)
	rng, err := util.NewRange(start, end)
	if err != nil {
		return fmt.Errorf("%w: %s", types.ErrBadArgument, err)
	}
	for ; iter.Valid(); iter.Next() {
		inRange, stopIter := rng.CheckAndMoveForward()
		if stopIter {
			break
		}
		if !inRange {
			continue
		}
		if !do(iter.Key()) {
			break
		}
	}
	iter.Close()
	return nil
}

// mapKey maps the given primary key to the secondary key, so when iterating a prefixed store
// created from a secondary key we will find the provided primary key.
func (s Store) mapKey(secondaryKey types.SecondaryKey, primaryKey []byte) (computedKey []byte, err error) {
	store, computedKey, err := s.kvStore(secondaryKey)
	if err != nil {
		return nil, err
	}
	if store.Has(primaryKey) {
		return nil, fmt.Errorf("%w: primary key %x in index %s", types.ErrAlreadyExists, primaryKey, secondaryKey)
	}
	store.Set(primaryKey, []byte{})
	return computedKey, nil
}

func (s Store) unmapRawKeys(primaryKey []byte, encodedKeys [][]byte) error {
	for _, encKey := range encodedKeys {
		store := s.kvStoreRaw(encKey)
		if !store.Has(primaryKey) {
			return fmt.Errorf("%w: key %x was not found in index key prefixed store %x", types.ErrNotFound, primaryKey, encKey)
		}
		store.Delete(primaryKey)
	}
	return nil
}

// kvStore returns the prefixed key value store from the given secondary key
// the secondary key is encoded in a way that it's impossible to iterate over
// domains of keys that have the provided secondary key as prefix
// the computed (encoded) key is returned for convenience as it's usually
// used to be saved in index keys list!
func (s Store) kvStore(sk types.SecondaryKey) (store sdk.KVStore, computedKey []byte, err error) {
	// compute key
	computedKey, err = encodeIndexKey(sk)
	if err != nil {
		return
	}
	// get prefixed store from the raw encoded secondary key
	store = s.kvStoreRaw(computedKey)
	return
}

// kvStoreRaw returns the prefixed key value store from an encoded types.SecondaryKey
func (s Store) kvStoreRaw(encodedKey []byte) sdk.KVStore {
	return prefix.NewStore(s.indexes, encodedKey)
}

// saveIndexList stores the encoded secondary keys the primary key
// is using, so in case the object needs to be updated or deleted
// it's fairly easy to understand which secondary keys point to it
// and so update or delete them
func (s Store) saveIndexList(primaryKey []byte, encodedKeys [][]byte) error {
	// sort keys deterministically
	util.SortByteSlice(encodedKeys)
	// marshal index list
	b, err := s.cdc.MarshalBinaryLengthPrefixed(&types.IndexList{Indexes: encodedKeys})
	if err != nil {
		return err
	}
	// save data to store
	if s.primaryKeysIndexes.Has(primaryKey) {
		return fmt.Errorf("%w: key %x already exists in index list store", types.ErrAlreadyExists, primaryKey)
	}
	s.primaryKeysIndexes.Set(primaryKey, b)
	return nil
}

// deleteIndexList deletes the secondary keys list of the given primary key
func (s Store) deleteIndexList(primaryKey []byte) error {
	if !s.primaryKeysIndexes.Has(primaryKey) {
		return fmt.Errorf("%w: key %x in index list store", types.ErrNotFound, primaryKey)
	}
	s.primaryKeysIndexes.Delete(primaryKey)
	return nil
}

// getIndexList returns the encoded secondary keys which point to the given primary key
func (s Store) getIndexList(primaryKey []byte) (indexes [][]byte, err error) {
	b := s.primaryKeysIndexes.Get(primaryKey)
	if b == nil {
		return nil, fmt.Errorf("%w: key %x not found in index list store", types.ErrNotFound, primaryKey)
	}
	list := new(types.IndexList)
	err = s.cdc.UnmarshalBinaryLengthPrefixed(b, list)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to unmarshal: %s", types.ErrInternal, err.Error())
	}
	return list.Indexes, nil
}
