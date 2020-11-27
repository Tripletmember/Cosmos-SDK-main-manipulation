package indexes

import (
	"bytes"
	"errors"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/test"
	"testing"
)

func TestStore(t *testing.T) {
	t.Run("success", makeTest(t, func(t *testing.T, store Store) {
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
		// delete
		err = store.Delete(obj.PrimaryKey())
		if err != nil {
			t.Fatal(err)
		}
		// ensure indexes are deleted
		pks, err = store.QueryAll(obj.SecondaryKeys()[0])
		if err != nil {
			t.Fatal(err)
		}
		if len(pks) != 0 {
			t.Fatalf("found unexpected primary keys: %v", pks)
		}
		// ensure index list is deleted too
		err = store.deleteIndexList(obj.PrimaryKey())
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatalf("unexpected error: %s", err)
		}
	}))
}
