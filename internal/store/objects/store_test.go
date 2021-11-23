package objects

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crud "github.com/iov-one/cosmos-sdk-crud"
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
		checkObject(t, &store, &obj)

		// test can't create object with same primary key twice
		err = store.Create(obj)
		if !errors.Is(err, crud.ErrAlreadyExists) {
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
		checkObject(t, &store, &obj)

		// test object not found with mutated key

		var expected test.Object
		err = store.Read(test.MutateBytes(obj.PrimaryKey()), expected)
		if !errors.Is(err, crud.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}

		// test nil primary key
		err = store.Read(nil, expected)
		if !errors.Is(err, crud.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}

		// test empty primary key
		err = store.Read(make([]byte, 0), expected)
		if !errors.Is(err, crud.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}

	})
	t.Run("update", func(t *testing.T) {
		// TODO: add testing of updating an object with modified pk (should this fail ? should this just update the
		// newly referenced object gracefully ?)

		// test object not found
		obj := test.NewRandomObject()
		err := store.Update(obj)
		if !errors.Is(err, crud.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
		// create object then update
		err = store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}

		// update with no changes
		err = store.Update(obj)
		if err != nil {
			t.Fatal(err)
		}

		// check if everything is still ok
		checkObject(t, &store, &obj)

		obj.TestSecondaryKeyA = []byte("test2")
		err = store.Update(obj)
		if err != nil {
			t.Fatal(err)
		}
		// check if it was updated correctly
		checkObject(t, &store, &obj)
	})

	t.Run("delete", func(t *testing.T) {
		// test not found
		err := store.Delete([]byte("does-not-exist"))
		if !errors.Is(err, crud.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
		// create arbitrary object
		obj := test.NewRandomObject()
		err = store.Create(obj)
		if err != nil {
			t.Fatal(err)
		}
		// delete non existing object
		err = store.Delete(test.MutateBytes(obj.PrimaryKey()))
		if !errors.Is(err, crud.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}

		// delete object
		err = store.Delete(obj.PrimaryKey())
		if err != nil {
			t.Fatal(err)
		}
		// try to get object
		var expected = test.NewObject()
		err = store.Read(obj.PrimaryKey(), expected)
		if !errors.Is(err, crud.ErrNotFound) {
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

func checkKeys(t *testing.T, actual [][]byte, objects []crud.Object) {
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

func createStoreWithRandomObjects(cdc codec.Codec, db sdk.KVStore, t *testing.T, n int, uniqueID string) (Store, []crud.Object) {
	store := NewStore(cdc, prefix.NewStore(db, []byte(uniqueID)))
	addToStore := func(obj crud.Object) error {
		return store.Create(obj)
	}
	return store, test.CreateRandomObjects(addToStore, t, n)
}

// Helpers functions for testing
func checkObject(t *testing.T, store *Store, expected *test.Object) {

	var actual = test.NewObject()
	var err = store.Read(expected.PrimaryKey(), actual)
	if err != nil {
		t.Fatal(err)
	}
	if err := actual.Equals(expected); err != nil {
		t.Fatal(err)
	}
}
