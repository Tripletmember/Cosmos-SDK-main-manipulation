package indexes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/util"
)

func (s Store) Filter(secondaryKeys []types.SecondaryKey, start, end uint64) ([][]byte, error) {

	if len(secondaryKeys) == 0 {
		return nil, types.ErrBadArgument
	}

	results := make([][]byte, 0)
	addToResults := func(key []byte) bool {
		results = append(results, key)
		return true
	}

	err := s.FilterWithCallback(secondaryKeys, start, end, addToResults)
	return results, err
}

func (s Store) FilterWithCallback(
	sks []types.SecondaryKey,
	start, end uint64,
	do func(primaryKey []byte) (stop bool),
) (err error) {
	rng, err := util.NewRange(start, end)
	if err != nil {
		return types.ErrBadArgument
	}

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
	for {
		pk, noMoreValues := moveForward(indexStores)
		inRange, stopIter := rng.CheckAndMoveForward()
		// if filtering over
		if noMoreValues || stopIter {
			break
		}
		// if we are in the range [start, end[
		if inRange {
			// if after do the caller wants to stop
			if !do(pk) {
				break
			}
		}
	}
	return nil
}

// moveForward takes the next key that is present in all the iterators result sets
// It assumes keys are ordered (byte-wise, as per bytes.Compare computes) in ascending order
// If the stop return value is false, then primaryKey is meaningless and there are no more results
func moveForward(iters []sdk.Iterator) (primaryKey []byte, stop bool) {
	// If no iterator is given, then there is no matching key
	n := len(iters)
	if n == 0 {
		return nil, true
	}

	// Initialisation: retrieve the first value of the first iterator
	if !iters[0].Valid() {
		return nil, true
	}
	// The current candidate key
	candidateKey := nextKey(iters[0])
	// The index of the tested iterator in the iterator array
	i := 1
	// The number of remaining iterators to check in order to validate the candidate key
	remainingIterators := n - 1

	// While we have not validated the key against all the iterators, continue
	for remainingIterators > 0 {
		// If any of the iterator is fully consumed then no more results can be found
		if !iters[i].Valid() {
			return nil, true
		}

		// We retrieve the next key from this iterator to test it against the candidate key
		currentKey := nextKey(iters[i])
		bigger, equal := util.BytesBiggerEqual(currentKey, candidateKey)
		// If the current key is greater than the candidate key, then there is no chance of validating the candidate key.
		// That is because, as this iterator is ordered in ascending order, we now can guarantee that the candidate key
		// is not present in this iterator's result set
		if bigger {
			// The current key is our new candidate key
			// We reset the remainingIterators counter and skip to the next iterator
			remainingIterators = n - 1
			candidateKey = currentKey
			i = (i + 1) % n
		} else if equal {
			// If the candidate key and the current key match, we decrement the remainingIterators counter
			// and skip to the next iterator
			remainingIterators--
			i = (i + 1) % n
		}
	}
	// If we reach the end of the function, the candidate key has been validated for all the iterators so is a valid key
	return candidateKey, false
}

// nextKey gets the key of the iterator and then moves it forward
func nextKey(it sdk.Iterator) []byte {
	key := it.Key()
	it.Next()
	return key
}
