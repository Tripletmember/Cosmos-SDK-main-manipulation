package indexes

import (
	"errors"
	"reflect"
	"testing"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
)

func Test_filtering(t *testing.T) {
	ctx, key, cdc, err := test.New()
	if err != nil {
		t.Fatalf("failed to create tests: %s", err)
	}
	testKVStore := ctx.KVStore(key)
	store := NewStore(cdc, testKVStore)

	// Add some test objects to the index in order to test filtering
	err = store.Index(test.NewCustomObject("pk4", "a2", "b3"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk5", "a3", "b3"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk2", "a2", "b2"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk3", "a1", "b2"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk1", "a1", "b1"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk90", "a2", "b3"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk7", "b1", "a1"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk6", "a4", "b3"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk8", "a4", "a4"))
	checkError(t, err)
	err = store.Index(test.NewCustomObject("pk9", "a21", "b21"))
	checkError(t, err)

	t.Run("empty sk set", func(t *testing.T) {
		_, err := store.Filter(make([]types.SecondaryKey, 0), 0, 0)
		if !errors.Is(err, types.ErrBadArgument) {
			t.Fatal("Unexpected error", err, "(expecting bad argument)")
		}
	})

	t.Run("single sk set", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a2")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk2", "pk4", "pk90"})
	})
	t.Run("single set w/ range limit", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 0, 2)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk4", "pk5"})
	})
	t.Run("single set w/ range offset", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 1, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk5", "pk6", "pk90"})
	})

	t.Run("single set w/ range limit and offset", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 1, 3)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk5", "pk6"})
	})

	t.Run("multiple sk set", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a4")},
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk6"})
	})
	t.Run("duplicated sk key", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a2")},
			{ID: 0x0, Value: []byte("a3")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})

	t.Run("duplicated sk key/value", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a2")},
			{ID: 0x0, Value: []byte("a2")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk2", "pk4", "pk90"})
	})
	t.Run("nonexistent sk key", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x4, Value: []byte("a2")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("nonexistent sk value", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("b2")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("multiple sk/empty result for 1st", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a11")},
			{ID: 0x1, Value: []byte("b1")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("multiple sk/empty result for 2nd", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x1, Value: []byte("b21")},
			{ID: 0x0, Value: []byte("a11")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("multiple sk/no results", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a1")},
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("multiple sk/empty result for last", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a4")},
			{ID: 0x0, Value: []byte("a4")},
			{ID: 0x1, Value: []byte("a4")},
			{ID: 0x1, Value: []byte("b4")},
		}
		pks, err := store.Filter(sks, 0, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("multiple sk/range", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a2")},
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 1, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk90"})
	})
	t.Run("multiple sk/empty result range", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a2")},
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 2, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
	t.Run("invalid range", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a1")},
		}
		_, err := store.Filter(sks, 1, 1)
		if !errors.Is(err, types.ErrBadArgument) {
			t.Fatal("Unexpected error", err, "(expecting bad argument)")
		}
		_, err = store.Filter(sks, 5, 1)
		if !errors.Is(err, types.ErrBadArgument) {
			t.Fatal("Unexpected error", err, "(expecting bad argument)")
		}
	})
	t.Run("infinite range w/ offset", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x1, Value: []byte("b3")},
		}
		pks, err := store.Filter(sks, 1, 0)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk5", "pk6", "pk90"})
	})

	t.Run("end index too far", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a1")},
		}
		pks, err := store.Filter(sks, 0, 25)
		checkError(t, err)
		checkExpected(t, pks, []string{"pk1", "pk3"})
	})

	t.Run("start index too far", func(t *testing.T) {
		sks := []types.SecondaryKey{
			{ID: 0x0, Value: []byte("a1")},
		}
		pks, err := store.Filter(sks, 22, 25)
		checkError(t, err)
		checkExpected(t, pks, []string{})
	})
}

func checkExpected(t *testing.T, actual [][]byte, expected []string) {
	expectedBytes := make([][]byte, len(expected))
	for i, val := range expected {
		expectedBytes[i] = []byte(val)
	}

	if actual == nil {
		t.Fatal("Nil slice returned")
	}

	if !reflect.DeepEqual(actual, expectedBytes) {
		t.Fatal("Result set does not match (expected :", expectedBytes, ", actual :", actual, ")")
	}
}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unexpected error : ", err)
	}

}
