package crud

import (
	"errors"
	"reflect"
	"testing"

	"github.com/iov-one/cosmos-sdk-crud/internal/store"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
)

func Test_query(t *testing.T) {
	// instantiate store and query
	db, cdc, err := test.NewStore()
	if err != nil {
		t.Fatal(err)
	}
	crudStore := store.NewStore(cdc, db, nil)
	var q = newQuery(crudStore)
	// apply changes
	var sk1TestValue = []byte("test")
	var sk2TestValue = []byte("test2")
	q.
		Index(0x0).
		Equals(sk1TestValue).
		And().
		Index(0x1).
		Equals(sk2TestValue).
		WithRange().Start(10).End(20)
	t.Run("success", func(t *testing.T) {
		_, err = q.Do()
		if err != nil {
			t.Fatal(err)
		}
		expected := &query{
			errs: nil,
			andEqualSk: map[byte]struct{}{
				0x0: {},
				0x1: {},
			},
			sks: []types.SecondaryKey{
				{
					ID:    0x0,
					Value: sk1TestValue,
				},
				{
					ID:    0x1,
					Value: sk2TestValue,
				},
			},
			currSk: types.SecondaryKey{
				ID:    0x1,
				Value: sk2TestValue,
			},
			store:    crudStore,
			start:    10,
			end:      20,
			consumed: true,
		}

		if !reflect.DeepEqual(expected, q) {
			t.Logf("%#v", expected)
			t.Logf("%#v", q)
			t.Fatal("unexpected result")
		}
	})

	t.Run("success/empty equality", func(t *testing.T) {
		q := newQuery(crudStore)
		_, err = q.Where().Index(0x0).Equals([]byte("")).Do()
		if err != nil {
			t.Fatal(err)
		}
		expected := &query{
			errs: nil,
			andEqualSk: map[byte]struct{}{
				0x0: {},
			},
			sks: []types.SecondaryKey{
				{
					ID:    0x0,
					Value: []byte(""),
				},
			},
			currSk: types.SecondaryKey{
				ID:    0x0,
				Value: []byte(""),
			},
			store:    crudStore,
			start:    0,
			end:      0,
			consumed: true,
		}

		if !reflect.DeepEqual(expected, q) {
			t.Logf("%#v", expected)
			t.Logf("%#v", q)
			t.Fatal("unexpected result")
		}
	})
	t.Run("bad argument/already consumed", func(t *testing.T) {
		_, _ = q.Do() // do it twice in case we run this subtest only!
		_, err := q.Do()
		t.Logf("%s", err)
		if !errors.Is(err, ErrBadArgument) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("bad argument/no secondary keys", func(t *testing.T) {
		q := newQuery(crudStore)
		_, err := q.Do()
		t.Logf("%s", err)
		if !errors.Is(err, ErrBadArgument) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("bad argument/multiple indexes with same id", func(t *testing.T) {
		q := newQuery(crudStore)
		q.Where().Index(0x1).Equals([]byte("1")).And().Index(0x1).Equals([]byte{0x1})
		_, err := q.Do()
		t.Logf("%s", err)
		if !errors.Is(err, ErrBadArgument) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("bad argument/nil equality", func(t *testing.T) {
		q := newQuery(crudStore)
		q.Where().Index(0x1).Equals(nil)
		_, err := q.Do()
		t.Logf("%s", err)
		if !errors.Is(err, ErrBadArgument) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

}
