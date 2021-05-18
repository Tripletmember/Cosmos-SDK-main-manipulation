package objects

import (
	"bytes"
	"errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sort"
	"testing"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
)

func TestStore(t *testing.T) {
	db, cdc, err := test.NewStore()
	if err != nil {
		t.Fatal("failed precondition", err)
	}
	store := NewStore(cdc, db)
	t.Run("create", func(t *testing.T) {
		obj := test.NewRandomObject()
		// test creation
		err := store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}
		// test correct unmarshalling
		var expected = test.NewObject()
		err = store.Read(obj.PrimaryKey(), expected)
		if err != nil {
			t.Fatal(err)
		}
		if err := obj.Equals(expected); err != nil {
			t.Fatal(err)
		}
		// test can't create object with same primary key twice
		err = store.Create(obj)
		if !errors.Is(err, types.ErrAlreadyExists) {
			t.Fatal("unexpected error", err)
		}
	})
	t.Run("read", func(t *testing.T) {
		obj := test.NewRandomObject()
		err := store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}
		// test correct unmarshalling
		var expected = test.NewObject()
		err = store.Read(obj.PrimaryKey(), expected)
		if err != nil {
			t.Fatal(err)
		}
		if err := obj.Equals(expected); err != nil {
			t.Fatal(err)
		}
		// test object not found
		err = store.Read(test.NewRandomObject().PrimaryKey(), expected)
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
	})
	t.Run("update", func(t *testing.T) {
		// test object not found
		obj := test.NewRandomObject()
		err := store.Update(obj)
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
		// create object then update
		err = store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}
		obj.TestSecondaryKeyA = []byte("test2")
		err = store.Update(obj)
		if err != nil {
			t.Fatal(err)
		}
		// check if it was updated correctly
		var expected = test.NewObject()
		err = store.Read(obj.PrimaryKey(), expected)
		if err != nil {
			t.Fatal(err)
		}
		if err := obj.Equals(expected); err != nil {
			t.Fatal(err)
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
		err = store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}
		// delete object
		err = store.Delete(obj.PrimaryKey())
		if err != nil {
			t.Fatal(err)
		}
		// try to get object
		var expected = test.NewObject()
		err = store.Read(obj.PrimaryKey(), expected)
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("get all key", func(t *testing.T) {
		store, objs := createStoreWithRandomObjects(cdc, db, t, 10, "allkey")

		actual, err := store.GetAllKeys(0, 0)
		if err != nil {
			t.Fatal("Unexpected error :", err)
		}
		checkKeys(t, actual, objs)
	})

	t.Run("get key range", func(t *testing.T) {
		store, objs := createStoreWithRandomObjects(cdc, db, t, 10, "keyrange")

		//TODO: uncomment this test when merging with the newer version of Range
		/*actual, err := store.GetAllKeys(5, 0)
		if err != nil {
			t.Fatal("Unexpected error :", err)
		}
		checkKeys(t, actual, objs[5:])*/

		actual, err := store.GetAllKeys(5, 7)
		if err != nil {
			t.Fatal("Unexpected error :", err)
		}
		checkKeys(t, actual, objs[5:7])

		actual, err = store.GetAllKeys(2, 12)
		if err != nil {
			t.Fatal("Unexpected error :", err)
		}
		checkKeys(t, actual, objs[2:])
	})
}

func checkKeys(t *testing.T, actual [][]byte, objects []types.Object) {
	if len(actual) != len(objects) {
		t.Fatalf("Result set length mismatch : actual = %v, expected = %v", len(actual), len(objects))
	}

	for i := 0; i < len(actual); i++ {
		expected := objects[i].PrimaryKey()
		if !bytes.Equal(actual[i], expected) {
			t.Fatalf("Invalid key at position %v : actual = %v, expected = %v", i, actual[i], expected)
		}
	}
}

func createStoreWithRandomObjects(cdc codec.Marshaler, db sdk.KVStore, t *testing.T, n int, uniqueID string) (Store, []types.Object) {
	store := NewStore(cdc, prefix.NewStore(db, []byte(uniqueID)))
	var objs []types.Object = nil

	for i := 0; i < n; i++ {
		obj := test.NewRandomObject()
		objs = append(objs, obj)
		err := store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}
	}

	sort.Slice(objs, func(i, j int) bool { return bytes.Compare(objs[i].PrimaryKey(), objs[j].PrimaryKey()) < 0 })
	return store, objs
}
