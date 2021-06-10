package indexes

import (
	"encoding/binary"
	"fmt"
	"math"

	types2 "github.com/iov-one/cosmos-sdk-crud/types"
)

// maxKeyLength defines the index key maximum length in bytes
const maxKeyLength = math.MaxUint16

// numBytesKeyLength defines the bytes required to express the length of a key
const numBytesKeyLength = 2

// encodeIndexKey takes a crud.SecondaryKey and encodes it
// the way in which it's encoded is the following
// key = <[1]byte=secondaryKey.ID><[2]byte=littleEndian(len(secondaryKey.Value))><[]byte(SecondaryKey.Value)>
// in this way we have keys, that when iterated, do not go over domains of longer keys which contain
// the key, example:
// keyA = <0x1,0x2,0x3>
// keyB = <0x1,0x2,0x3,0x4>
// if we wanted to iterate over index keyA we would end up in keyB domain too
// as keyB has, as prefix, keyA. In order to avoid this we encode keys using
// the strategy highlighted above.
// This functions treats a nil secondary key value as an empty value, and thus a nil value will be transformed into
// an empty byte array through encode-decode
// Error types are of types.ErrBadArgument, and happen when
// the representation of the length of the index key takes more than 2 bytes (uint16).
func encodeIndexKey(sk types2.SecondaryKey) ([]byte, error) {
	length := len(sk.Value)
	if length > maxKeyLength {
		return nil, fmt.Errorf("%w: index keys bigger than %d bytes are not allowed, got: %d", types2.ErrBadArgument, maxKeyLength, length)
	}
	encodedLength := make([]byte, numBytesKeyLength)
	binary.LittleEndian.PutUint16(encodedLength, uint16(length))
	finalKey := append([]byte{byte(sk.ID)}, encodedLength...)
	return append(finalKey, sk.Value...), nil
}

// decodeIndexKey takes a key and tries to turn it into a secondary key
// error returned here are of types.ErrInternal, this is because
// since, before doing a decode, you must have done an encode
// and they should be respectively their reverse functions
// it means that either state was corrupted or there is no backwards compatibility
// between the two anymore.
func decodeIndexKey(key []byte) (sk types2.SecondaryKey, err error) {
	// minimumKeyLength defines the minimum length a key has to have
	// to be converted into a secondary key
	const metadataLength = numBytesKeyLength + 1
	// sanity checks
	length := len(key)
	if length < metadataLength {
		return sk, fmt.Errorf("%w: minimum length not reached, got: %d, want: %d", types2.ErrInternal, len(key), metadataLength)
	}
	decodedLength := binary.LittleEndian.Uint16(key[1:3])
	valueLength := length - metadataLength
	if int(decodedLength) != valueLength {
		return sk, fmt.Errorf("%w, mismatch in length, decoded: %d, got: %d", types2.ErrInternal, decodedLength, valueLength)
	}
	// create secondary key
	value := make([]byte, length-metadataLength)
	copy(value, key[metadataLength:])
	return types2.SecondaryKey{
		ID:    types2.IndexID(key[0]),
		Value: value,
	}, nil
}
