package indexes

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"testing"

	crud "github.com/iov-one/cosmos-sdk-crud"
)

func Test_encodeDecodeIndexKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sk := crud.SecondaryKey{
			ID:    0x0,
			Value: []byte("test"),
		}
		key, err := encodeIndexKey(sk)
		if err != nil {
			t.Fatal(err)
		}
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
		_, err := encodeIndexKey(crud.SecondaryKey{
			ID:    0x1,
			Value: c,
		})
		if !errors.Is(err, crud.ErrBadArgument) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("encode/prefix check", func(t *testing.T) {
		c, c2 := []byte("myKey"), []byte("myKey2")

		k, err := encodeIndexKey(crud.SecondaryKey{
			ID:    0x1,
			Value: c,
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		k2, err := encodeIndexKey(crud.SecondaryKey{
			ID:    0x1,
			Value: c2,
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if bytes.HasPrefix(k2, k) {
			t.Fatal("A key is a prefix of another")
		}
	})
	t.Run("encode/empty key", func(t *testing.T) {
		var ek1, ek2 []byte
		var dk1, dk2 crud.SecondaryKey
		var err error

		// Encode k1, a sk with empty value, into ek1
		k1 := crud.SecondaryKey{
			ID:    0x1,
			Value: make([]byte, 0),
		}
		ek1, err = encodeIndexKey(k1)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Encode k2, a sk with nil value, into ek2
		k2 := crud.SecondaryKey{
			ID:    0x1,
			Value: nil,
		}
		ek2, err = encodeIndexKey(k2)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Decode ek1 into dk1 and compare it to the original key k1
		dk1, err = decodeIndexKey(ek1)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !reflect.DeepEqual(k1, dk1) {
			t.Fatal("Invalid encode/decode on secondary key with empty value")
		}

		// Decode ek2 into dk2, nil values are transformed into empty value, so check that dk2 is equal to k1
		dk2, err = decodeIndexKey(ek2)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Nil becomes an empty byte array through encode/decode
		if !reflect.DeepEqual(k1, dk2) {
			t.Fatal("Invalid encode/decode on secondary key with nil value")
		}

	})
	t.Run("decode/error minimum length", func(t *testing.T) {
		c := make([]byte, 2)
		_, err := decodeIndexKey(c)
		if !errors.Is(err, crud.ErrInternal) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
	t.Run("decode/error length too small", func(t *testing.T) {
		key := []byte{
			0x0, // byte: id byte
			0x0, // byte: length byte 0
			0x0, // byte: length byte 1
			0x1, // byte: key 1
		}
		_, err := decodeIndexKey(key)
		if !errors.Is(err, crud.ErrInternal) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("decode/error length too big", func(t *testing.T) {
		key := []byte{
			0x0, // byte: id byte
			0x0, // byte: length byte 0
			0x2, // byte: length byte 1
			0x1, // byte: key 1
		}
		_, err := decodeIndexKey(key)
		if !errors.Is(err, crud.ErrInternal) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
}
