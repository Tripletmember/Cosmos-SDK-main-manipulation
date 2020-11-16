package indexes

import (
	"github.com/fdymylja/cosmos-sdk-oodb/internal/test"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func makeTest(_ *testing.T, testFn func(t *testing.T, store Store)) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, key, cdc, err := test.New()
		if err != nil {
			t.Fatalf("failed to create tests: %s", err)
		}
		testKVStore := ctx.KVStore(key)
		testStore := NewStore(cdc, testKVStore)
		testFn(t, testStore)
	}
}
