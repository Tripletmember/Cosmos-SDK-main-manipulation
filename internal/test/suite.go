package test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crud "github.com/iov-one/cosmos-sdk-crud"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
)

// Key is the name of a test key
const Key = "test"
const IndexID_A = 0x0
const IndexID_B = 0x1

// assert crud.Object is implemented by test Object
var _ = crud.Object(NewObject())

// NewStore builds a new store
func NewStore() (sdk.KVStore, codec.Marshaler, error) {
	ctx, storeKey, cdc, err := New()
	if err != nil {
		return nil, nil, err
	}
	return ctx.KVStore(storeKey), cdc, nil
}

// New returns the objects necessary to run a test
func New() (sdk.Context, sdk.StoreKey, codec.Marshaler, error) {
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("crud.internal.test",
		(*crud.Object)(nil),
		&Object{},
	)
	testCdc := codec.NewProtoCodec(interfaceRegistry)
	testKey := sdk.NewKVStoreKey(Key)
	mdb := db.NewMemDB()
	ms := store.NewCommitMultiStore(mdb)
	ms.MountStoreWithDB(testKey, sdk.StoreTypeIAVL, mdb)
	err := ms.LoadLatestVersion()
	if err != nil {
		return sdk.Context{}, nil, nil, err
	}
	testCtx := sdk.NewContext(ms, tmproto.Header{Time: time.Now()}, true, log.NewNopLogger())
	return testCtx, testKey, testCdc, nil
}

func NewCustomObject(pk, sk1, sk2 string) Object {
	return Object{
		TestObject: &types.TestObject{
			TestPrimaryKey:    []byte(pk),
			TestSecondaryKeyA: []byte(sk1),
			TestSecondaryKeyB: []byte(sk2),
		},
	}
}

func NewDeterministicObject() Object {
	pk := []byte("primary-key")
	skA := []byte("secondary-key")
	skB := []byte("secondary-key1")
	return Object{
		TestObject: &types.TestObject{
			TestPrimaryKey:    pk,
			TestSecondaryKeyA: skA,
			TestSecondaryKeyB: skB,
		},
	}
}

func NewRandomObject() Object {
	pk := make([]byte, 8)
	_, err := rand.Read(pk)
	if err != nil {
		panic(err)
	}
	skA := make([]byte, 8)
	_, err = rand.Read(skA)
	if err != nil {
		panic(err)
	}
	skB := make([]byte, 8)
	_, err = rand.Read(skB)
	if err != nil {
		panic(err)
	}
	return Object{
		&types.TestObject{
			TestPrimaryKey:    pk,
			TestSecondaryKeyA: skA,
			TestSecondaryKeyB: append(skA, skB...),
		},
	}
}

type Object struct {
	*types.TestObject
}

func NewObject() *Object {
	testObject := types.TestObject{}
	object := Object{&testObject}

	return &object
}

func (o Object) PrimaryKey() (primaryKey []byte) {
	return o.TestPrimaryKey
}

func (o Object) SecondaryKeys() (secondaryKeys []crud.SecondaryKey) {
	return []crud.SecondaryKey{
		{
			ID:    IndexID_A,
			Value: o.TestSecondaryKeyA,
		},
		{
			ID:    IndexID_B,
			Value: o.TestSecondaryKeyB,
		},
	}
}

func (o Object) FirstSecondaryKey() crud.SecondaryKey {
	return crud.SecondaryKey{
		ID:    IndexID_A,
		Value: o.TestSecondaryKeyA,
	}
}

func (o Object) SecondSecondaryKey() crud.SecondaryKey {
	return crud.SecondaryKey{
		ID:    IndexID_B,
		Value: o.TestSecondaryKeyB,
	}
}

func (this *Object) Equals(that *Object) error {
	tester := func(a []uint8, b []uint8) error {
		if len(a) != len(b) {
			return fmt.Errorf("len(a) == %d != len(b) == %d", len(a), len(b))
		}
		for i, ai := range a {
			if ai != b[i] {
				return fmt.Errorf("a[%d] == %d != b[%d] == %d", i, ai, i, b[i])
			}
		}
		return nil
	}

	if err := tester(this.TestPrimaryKey, that.TestPrimaryKey); err != nil {
		return err
	}
	if err := tester(this.TestSecondaryKeyA, that.TestSecondaryKeyA); err != nil {
		return err
	}
	if err := tester(this.TestSecondaryKeyB, that.TestSecondaryKeyB); err != nil {
		return err
	}

	return nil
}

func MutateBytes(originalBytes []byte) (mutatedBytes []byte) {
	if originalBytes != nil {
		mutatedBytes = make([]byte, len(originalBytes))
		copy(mutatedBytes, originalBytes)
		mutatedBytes[len(mutatedBytes)-1] += 1
	}
	return
}

func CreateRandomObjects(add func(crud.Object) error, t *testing.T, n int) []crud.Object {
	var objs []crud.Object = nil

	for i := 0; i < n; i++ {
		obj := NewRandomObject()
		objs = append(objs, obj)
		err := add(obj)
		if err != nil {
			t.Fatal(err)
		}
	}

	sort.Slice(objs, func(i, j int) bool { return bytes.Compare(objs[i].PrimaryKey(), objs[j].PrimaryKey()) < 0 })
	return objs
}

func CheckNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unexpected error : ", err)
	}

}
