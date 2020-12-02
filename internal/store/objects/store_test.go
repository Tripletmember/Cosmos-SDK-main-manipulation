package objects

import (
	"errors"
	"reflect"
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
		var expected = test.Object{}
		err = store.Read(obj.PrimaryKey(), &expected)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(obj, expected) {
			t.Fatal("unexpected result")
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
		var expected = test.Object{}
		err = store.Read(obj.PrimaryKey(), &expected)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(obj, expected) {
			t.Fatal("unexpected result")
		}
		// test object not found
		err = store.Read(test.NewRandomObject().PrimaryKey(), &expected)
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
		obj.ProtobufObject.TestSecondaryKeyA = []byte("test2")
		err = store.Update(obj)
		if err != nil {
			t.Fatal(err)
		}
		// check if it was updated correctly
		var expected = test.Object{}
		err = store.Read(obj.PrimaryKey(), &expected)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, obj) {
			t.Fatal("unexpected result")
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
		var expected = test.Object{}
		err = store.Read(obj.PrimaryKey(), &expected)
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal("unexpected error", err)
		}
	})
}
