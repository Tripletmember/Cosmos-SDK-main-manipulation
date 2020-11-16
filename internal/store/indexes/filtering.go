package indexes

import (
	"fmt"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/store/types"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/util"
	"sort"
)

func (s Store) Filter(secondaryKeys []types.SecondaryKey, start, end uint64) ([][]byte, error) {
	// create rng
	rng, err := util.NewRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", types.ErrBadArgument, err)
	}
	var sets []util.ByteSet
	for _, sk := range secondaryKeys {
		set := util.NewByteSet()
		err := s.Query(sk, 0, 0, func(primaryKey []byte) (stop bool) {
			set.Insert(primaryKey)
			return false
		})
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}
	// filter and find common primary keys
	return findCommon(sets, rng), nil
}

func findCommon(sets []util.ByteSet, rng *util.Range) [][]byte {
	// order from smallest to biggest
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].Len() < sets[j].Len()
	})
	var pks [][]byte
	for _, k := range sets[0].Range() {
		if !isInAll(k, sets[1:]) {
			continue
		}
		inRange, stopIter := rng.CheckAndMoveForward()
		if stopIter {
			break
		}
		if !inRange {
			continue
		}
		pks = append(pks, k)
	}
	return pks
}

func isInAll(key []byte, sets []util.ByteSet) bool {
	for _, set := range sets {
		if !set.Has(key) {
			return false
		}
	}
	return true
}
