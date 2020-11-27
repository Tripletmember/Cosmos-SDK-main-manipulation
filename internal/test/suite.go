package test

import (
	"crypto/rand"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	tmtypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
	"time"
)

// Key is the name of a test key
const Key = "test"
const IndexID_A = 0x0
const IndexID_B = 0x1

// assert types.Object is implemented by test Object
var _ = types.Object(Object{})

// NewStore builds a new store
func NewStore() (sdk.KVStore, *codec.Codec, error) {
	ctx, storeKey, cdc, err := New()
	if err != nil {
		return nil, nil, err
	}
	return ctx.KVStore(storeKey), cdc, nil
}

// New returns the objects necessary to run a test
func New() (sdk.Context, sdk.StoreKey, *codec.Codec, error) {
	testCdc := codec.New()
	testKey := sdk.NewKVStoreKey(Key)
	mdb := db.NewMemDB()
	ms := store.NewCommitMultiStore(mdb)
	ms.MountStoreWithDB(testKey, sdk.StoreTypeIAVL, mdb)
	err := ms.LoadLatestVersion()
	if err != nil {
		return sdk.Context{}, nil, nil, err
	}
	testCtx := sdk.NewContext(ms, tmtypes.Header{Time: time.Now()}, true, log.NewNopLogger())
	return testCtx, testKey, testCdc, nil
}

func NewDeterministicObject() Object {
	pk := []byte("primary-key")
	skA := []byte("secondary-key")
	skB := []byte("secondary-key1")
	return Object{
		TestPrimaryKey:    pk,
		TestSecondaryKeyA: skA,
		TestSecondaryKeyB: skB,
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
		TestPrimaryKey:    pk,
		TestSecondaryKeyA: skA,
		TestSecondaryKeyB: append(skA, skB...),
	}
}

// Object is a mock object used to test the store
type Object struct {
	// TestPrimaryKey is a primary key
	TestPrimaryKey []byte
	// TestSecondaryKeyA is secondary key number one
	TestSecondaryKeyA []byte
	// TestSecondaryKeyB is secondary key number two
	TestSecondaryKeyB []byte
}

func (o Object) PrimaryKey() (primaryKey []byte) {
	return o.TestPrimaryKey
}

func (o Object) SecondaryKeys() (secondaryKeys []types.SecondaryKey) {
	return []types.SecondaryKey{
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

func (o Object) FirstSecondaryKey() types.SecondaryKey {
	return types.SecondaryKey{
		ID:    IndexID_A,
		Value: o.TestSecondaryKeyA,
	}
}

func (o Object) SecondSecondaryKey() types.SecondaryKey {
	return types.SecondaryKey{
		ID:    IndexID_B,
		Value: o.TestSecondaryKeyB,
	}
}

func (o Object) Reset() {
	o.TestSecondaryKeyA = nil
	o.TestPrimaryKey = nil
	o.TestSecondaryKeyB = nil
}
