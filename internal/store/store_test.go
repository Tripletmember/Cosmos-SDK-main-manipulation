package store

import (
	"errors"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/store/types"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/test"
	"reflect"
	"testing"
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
	var expected test.Object
	err = s.Read(obj.PrimaryKey(), &expected)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, obj) {
		t.Fatal("unexpected result")
	}
	// test update
	update := obj
	update.TestSecondaryKeyB = []byte("test-update")
	err = s.Update(update)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Read(obj.PrimaryKey(), &expected)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(update, expected) {
		t.Fatal("unexpected result")
	}
	// test cursor
	crs, err := s.Query([]types.SecondaryKey{
		update.FirstSecondaryKey(),
	}, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	// test read
	err = crs.Read(&expected)
	if err != nil {
		t.Logf("%s", crs.currKey())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, update) {
		t.Fatal("unexpected result")
	}
	// test update
	update.TestSecondaryKeyA = []byte("another-update")
	err = crs.Update(update)
	if err != nil {
		t.Fatal(err)
	}
	expected.Reset()
	err = crs.Read(&expected)
	if !reflect.DeepEqual(expected, update) {
		t.Fatal(err)
	}
	// test delete
	err = crs.Delete()
	if err != nil {
		t.Fatal(err)
	}
	err = crs.Read(&expected)
	if !errors.Is(err, types.ErrNotFound) {
		t.Fatal("unexpected error", err)
	}
}
