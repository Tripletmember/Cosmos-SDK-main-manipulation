package test

import (
	"crypto/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
)

// Key is the name of a test key
const Key = "test"
const IndexID_A = 0x0
const IndexID_B = 0x1

// assert types.Object is implemented by test Object
var _ = types.Object(Object{})

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
		(*types.Object)(nil),
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

func NewDeterministicObject() Object {
	pk := []byte("primary-key")
	skA := []byte("secondary-key")
	skB := []byte("secondary-key1")
	return Object{
		ProtobufObject: types.TestObject{
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
		ProtobufObject: types.TestObject{
			TestPrimaryKey:    pk,
			TestSecondaryKeyA: skA,
			TestSecondaryKeyB: append(skA, skB...),
		},
	}
}

type Object struct {
	types.Object

	ProtobufObject types.TestObject
}

func (o Object) Marshal() (bz []byte, err error) {
	return o.ProtobufObject.Marshal()
}

func (o Object) MarshalTo(bz []byte) (n int, err error) {
	return o.ProtobufObject.MarshalTo(bz)
}

func (o Object) MarshalToSizedBuffer(bz []byte) (int, error) {
	return o.ProtobufObject.MarshalToSizedBuffer(bz)
}

func (o Object) Size() (n int) {
	return o.ProtobufObject.Size()
}

func (o Object) Unmarshal(bz []byte) (err error) {
	// TODO: USEME return o.ProtobufObject.Unmarshal(bz)
	err = o.ProtobufObject.Unmarshal(bz)
	return err
}

func (o Object) PrimaryKey() (primaryKey []byte) {
	return o.ProtobufObject.TestPrimaryKey
}

func (o Object) SecondaryKeys() (secondaryKeys []types.SecondaryKey) {
	return []types.SecondaryKey{
		{
			ID:    IndexID_A,
			Value: o.ProtobufObject.TestSecondaryKeyA,
		},
		{
			ID:    IndexID_B,
			Value: o.ProtobufObject.TestSecondaryKeyB,
		},
	}
}

func (o Object) FirstSecondaryKey() types.SecondaryKey {
	return types.SecondaryKey{
		ID:    IndexID_A,
		Value: o.ProtobufObject.TestSecondaryKeyA,
	}
}

func (o Object) SecondSecondaryKey() types.SecondaryKey {
	return types.SecondaryKey{
		ID:    IndexID_B,
		Value: o.ProtobufObject.TestSecondaryKeyB,
	}
}

func (o Object) Reset() {
	o.ProtobufObject.TestSecondaryKeyA = nil
	o.ProtobufObject.TestPrimaryKey = nil
	o.ProtobufObject.TestSecondaryKeyB = nil
}
