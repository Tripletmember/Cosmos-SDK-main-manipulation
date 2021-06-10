package indexes

import (
	"bytes"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	crud "github.com/iov-one/cosmos-sdk-crud"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
)

func TestStore(t *testing.T) {

	store := createTestStore()

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
		if !errors.Is(err, crud.ErrAlreadyExists) {
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
		pks, err = store.QueryAll(crud.SecondaryKey{ID: 1, Value: make([]byte, 0)})
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
		if !errors.Is(err, crud.ErrNotFound) {
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
		if !errors.Is(err, crud.ErrNotFound) {
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
		if !errors.Is(err, crud.ErrNotFound) {
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

func BenchmarkStore_Index(b *testing.B) {
	s := createTestStore()

	var sk1, sk2 strings.Builder
	const bytesPerSK = 80
	for i := 0; i < bytesPerSK; i++ {
		sk1.WriteByte(byte(rand.Int()))
		sk2.WriteByte(byte(rand.Int()))
	}

	o := test.NewCustomObject("pk1", sk1.String(), sk2.String())

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		o.TestPrimaryKey = []byte("pk" + strconv.FormatInt(int64(n), 36))
		s.Index(o)
	}
}

// Helper functions for testing
func checkIndex(t *testing.T, store *Store, expected *test.Object) {
	var pks, err = store.QueryAll(expected.SecondaryKeys()[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(pks) == 1 && !bytes.Equal(pks[0], expected.PrimaryKey()) {
		t.Fatal("Primary key mismatch")
	}
}

func createTestStore() Store {
	ctx, key, cdc, err := test.New()
	if err != nil {
		panic("failed to create store:" + err.Error())
	}
	testKVStore := ctx.KVStore(key)
	return NewStore(cdc, testKVStore)
}
