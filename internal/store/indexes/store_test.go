package indexes

import (
	"bytes"
	"errors"
	"testing"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
)

func TestStore(t *testing.T) {

	ctx, key, cdc, err := test.New()
	if err != nil {
		t.Fatalf("failed to create tests: %s", err)
	}
	testKVStore := ctx.KVStore(key)
	store := NewStore(cdc, testKVStore)

	t.Run("index", func(t *testing.T) {
		obj := test.NewRandomObject()
		// test creation
		err := store.Index(obj)
		if err != nil {
			t.Fatal(err)
		}
		// test correct index
		checkIndex(t, &store, &obj)

		// test creation with same primary key twice
		err = store.Index(obj)
		if !errors.Is(err, types.ErrAlreadyExists) {
			t.Fatal("unexpected error", err)
		}
	})
	t.Run("read", func(t *testing.T) {
		obj := test.NewRandomObject()
		err := store.Index(obj)
		if err != nil {
			t.Fatal(err)
		}
		// test correct index
		checkIndex(t, &store, &obj)

		// test empty query
		var pks [][]byte
		pks, err = store.QueryAll(types.SecondaryKey{ID: 1, Value: make([]byte, 0)})
		if err != nil {
			t.Fatal("unexpected error", err)
		}

		if len(pks) != 0 {
			t.Fatal("Querying a nonexistent secondary key should result in an empty result set")
		}
	})

	t.Run("delete", func(t *testing.T) {
		// test not found
		err := store.Delete([]byte("does-not-exist"))
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
		// create arbitrary object
		obj := test.NewRandomObject()
		err = store.Index(obj)
		if err != nil {
			t.Fatal(err)
		}
		// delete non existing object
		err = store.Delete(test.MutateBytes(obj.PrimaryKey()))
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}

		// delete object
		err = store.Delete(obj.PrimaryKey())
		if err != nil {
			t.Fatal(err)
		}
		// try to get object
		var pks [][]byte
		pks, err = store.QueryAll(obj.SecondaryKeys()[1])
		if err != nil {
			t.Fatal("unexpected error", err)
		}
		// ensure that index has been deleted
		if len(pks) != 0 {
			t.Fatal("This object should have been deleted")
		}
		// ensure index list is deleted too
		err = store.deleteIndexList(obj.PrimaryKey())
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		obj := test.NewDeterministicObject()
		// index
		err := store.Index(obj)
		if err != nil {
			t.Fatal(err)
		}
		// query
		pks, err := store.QueryAll(obj.SecondaryKeys()[0])
		if err != nil {
			t.Fatal(err)
		}
		if len(pks) != 1 {
			t.Fatal("unexpected number of primary keys ", len(pks))
		}
		if !bytes.Equal(pks[0], obj.PrimaryKey()) {
			t.Fatalf("unexpected primary key: %x, wanted: %x", pks[0], obj.PrimaryKey())
		}
	})
}

// Helpers functions for testing
func checkIndex(t *testing.T, store *Store, expected *test.Object) {
	var pks, err = store.QueryAll(expected.SecondaryKeys()[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(pks) == 1 && !bytes.Equal(pks[0], expected.PrimaryKey()) {
		t.Fatal("Primary key mismatch")
	}
}
