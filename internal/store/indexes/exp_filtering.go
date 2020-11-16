// +build experimental_filtering

package indexes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/store/types"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/util"
)

func (s Store) experimentalFiltering(
	sks []types.SecondaryKey,
	start, end uint64,
	do func(primaryKey []byte) (stop bool),
) (err error) {
	indexStores := make([]sdk.Iterator, len(sks))
	var kv sdk.KVStore
	for i, sk := range sks {
		kv, _, err = s.kvStore(sk)
		if err != nil {
			return err
		}
		iter := kv.Iterator(nil, nil)
		indexStores[i] = iter
	}
	var i uint64 = 0
	for {
		pk, stop := moveForward(indexStores)
		// if filtering over
		if stop {
			break
		}
		// if out of range
		if i < start {
			continue
		}
		// if out of range
		if i != 0 && i > end {
			break
		}
		// if after do the caller wants to stop
		if !do(pk) {
			break
		}
		// move forward
		i++
	}
	return nil
}

func moveForward(iters []sdk.Iterator) (primaryKey []byte, stop bool) {
	if len(iters) == 0 {
		return nil, true
	}
	curr := []byte{}
	// move each iterator to the same level
	for _, iter := range iters {
		for {
			if !iter.Valid() {
				return nil, true
			}
			k := iter.Key()
			if util.BytesBiggerEqual(k, curr) {
				curr = k
				break
			}
			iter.Next()
		}
	}
	panic("not implemented")
}
