package store

import (
	"bytes"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
)

func TestStore(t *testing.T) {
	db, cdc, err := test.NewStore()
	if err != nil {
		t.Fatal("failed precondition", err)
	}
	s := NewStore(cdc, db, nil)
	obj := test.NewDeterministicObject()
	// test create
	err = s.Create(obj)
	if err != nil {
		t.Fatal(err)
	}
	// test read
	var expected = test.NewObject()
	err = s.Read(obj.PrimaryKey(), expected)
	if err != nil {
		t.Fatal(err)
	}
	if err := obj.Equals(expected); err != nil {
		t.Fatal(err)
	}
	// test update
	update := obj
	update.TestSecondaryKeyB = []byte("test-update")
	err = s.Update(update)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Read(obj.PrimaryKey(), expected)
	if err != nil {
		t.Fatal(err)
	}
	if err := update.Equals(expected); err != nil {
		t.Fatal(err)
	}
	// test cursor
	crs, err := s.Query([]types.SecondaryKey{
		update.FirstSecondaryKey(),
	}, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	// test read
	err = crs.Read(expected)
	if err != nil {
		t.Logf("%s", crs.currKey())
		t.Fatal(err)
	}
	if err := update.Equals(expected); err != nil {
		t.Fatal(err)
	}
	// test update
	update.TestSecondaryKeyA = []byte("another-update")
	err = crs.Update(update)
	if err != nil {
		t.Fatal(err)
	}
	expected.Reset()
	err = crs.Read(expected)
	if err := update.Equals(expected); err != nil {
		t.Fatal(err)
	}
	// test delete
	err = crs.Delete()
	if err != nil {
		t.Fatal(err)
	}
	err = crs.Read(expected)
	if !errors.Is(err, types.ErrNotFound) {
		t.Fatal("unexpected error", err)
	}

	t.Run("query all", func(t *testing.T) {
		s, objs := createStoreWithRandomObjects(cdc, db, t, 50, "queryall")

		results, err := s.Query(nil, 0, 0)
		if err != nil {
			t.Fatal("unexpected error", err)
		}

		i := 0
		for ; results.Valid(); results.Next() {
			if i == len(objs) {
				t.Fatalf("Length mismatch, exepected %v elements but got more", len(objs))
			}
			var actual = test.NewObject()
			if err := results.Read(*actual); err != nil {
				t.Fatal("Unexpected error :", err)
			}
			if !reflect.DeepEqual(*actual, objs[i]) {
				t.Fatalf("Object mismatch at index %v : expected = %v(%[2]T), actual = %v(%[3]T)", i, objs[i], actual)
			}
			i++
		}

	})
}

func createStoreWithRandomObjects(cdc codec.Marshaler, db sdk.KVStore, t *testing.T, n int, uniqueID string) (Store, []types.Object) {
	store := NewStore(cdc, db, []byte(uniqueID))
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
