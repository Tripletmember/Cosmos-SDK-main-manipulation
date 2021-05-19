package indexes

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
)

func Test_encodeDecodeIndexKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sk := types.SecondaryKey{
			ID:    0x0,
			Value: []byte("test"),
		}
		key, err := encodeIndexKey(sk)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", key)
		decoded, err := decodeIndexKey(key)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(sk, decoded) {
			t.Fatalf("unexpected result: want: %s, got: %s", sk, decoded)
		}
	})
	t.Run("encode/error max length", func(t *testing.T) {
		c := make([]byte, math.MaxUint16+1)
		_, err := encodeIndexKey(types.SecondaryKey{
			ID:    0x1,
			Value: c,
		})
		if !errors.Is(err, types.ErrBadArgument) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("decode/error minimum length", func(t *testing.T) {
		c := make([]byte, 1)
		_, err := decodeIndexKey(c)
		if !errors.Is(err, types.ErrInternal) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("decode/error length mismatch", func(t *testing.T) {
		key := []byte{
			0x0, // byte: length byte 0
			0x0, // byte: length byte 1
			0x0, // byte: id byte
			0x1, // byte: key 1
		}
		_, err := decodeIndexKey(key)
		if !errors.Is(err, types.ErrInternal) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
}
