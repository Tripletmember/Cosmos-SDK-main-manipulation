package crud

import (
	"errors"

	"github.com/iov-one/cosmos-sdk-crud/internal/store"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
)

// toExternalError converts a crud internal error to external
func toExternalError(err error) error {
	if err == nil {
		return nil
	}
	// check the internal error type
	if errors.Is(err, types.ErrAlreadyExists) {

	}
	if errors.Is(err, types.ErrBadArgument) {

	}
	if errors.Is(err, types.ErrCursorConsumed) {

	}
	if errors.Is(err, types.ErrInternal) {

	}
	if errors.Is(err, types.ErrNotFound) {

	}
	// does not match any known internal error
	return err
}

// toInternalOptions converts exported options to internal
func toInternalOptions(opts []OptionFunc) []store.OptionFunc {
	if len(opts) > 0 {
		panic("not implemented")
	}
	return nil
}

// toInternalObject converts the exported object to the internal one
func toInternalObject(o Object) types.Object {
	extSks := o.SecondaryKeys()
	sks := make([]types.SecondaryKey, len(extSks))
	for i, extSk := range extSks {
		sks[i] = types.SecondaryKey{
			ID:    byte(extSk.ID),
			Value: extSk.Value,
		}
	}
	return internalObjectWrapper{
		pk:                   o.PrimaryKey(),
		sks:                  sks,
		marshal:              o.Marshal,
		marshalTo:            o.MarshalTo,
		marshalToSizedBuffer: o.MarshalToSizedBuffer,
		size:                 o.Size,
		unmarshal:            o.Unmarshal,
		reset:                o.Reset,
		toString:             o.String,
		protoMessage:         o.ProtoMessage,
	}
}

type internalObjectWrapper struct {
	pk                   []byte
	sks                  []types.SecondaryKey
	marshal              func() ([]byte, error)
	marshalTo            func([]byte) (int, error)
	marshalToSizedBuffer func([]byte) (int, error)
	size                 func() int
	unmarshal            func([]byte) error
	reset                func()
	toString             func() string
	protoMessage         func()
}

func (i internalObjectWrapper) SecondaryKeys() []types.SecondaryKey {
	return i.sks
}

func (i internalObjectWrapper) PrimaryKey() []byte {
	return i.pk
}

func (i internalObjectWrapper) Marshal() (bz []byte, err error) {
	return i.marshal()
}

func (i internalObjectWrapper) MarshalTo(bz []byte) (n int, err error) {
	return i.marshalTo(bz)
}

func (i internalObjectWrapper) MarshalToSizedBuffer(bz []byte) (int, error) {
	return i.marshalToSizedBuffer(bz)
}

func (i internalObjectWrapper) Size() (n int) {
	return i.size()
}

func (i internalObjectWrapper) Unmarshal(bz []byte) (err error) {
	return i.unmarshal(bz)
}

func (i internalObjectWrapper) Reset() {
	i.reset()
}

func (i internalObjectWrapper) String() string {
	return i.toString()
}

func (i internalObjectWrapper) ProtoMessage() {
	i.protoMessage()
}

// storeWrapper wraps the internal store
type storeWrapper struct {
	s store.Store
}

func (i storeWrapper) Create(o Object) error {
	err := i.s.Create(toInternalObject(o))
	return toExternalError(err)
}

func (i storeWrapper) Read(primaryKey []byte, o Object) error {
	err := i.s.Read(primaryKey, toInternalObject(o))
	return toExternalError(err)
}

func (i storeWrapper) Update(o Object) error {
	err := i.s.Update(toInternalObject(o))
	return toExternalError(err)
}

func (i storeWrapper) Delete(primaryKey []byte) error {
	err := i.s.Delete(primaryKey)
	return toExternalError(err)
}

func (i storeWrapper) Query() QueryStatement {
	return newQuery(i.s)
}

type cursorWrapper struct {
	c *store.Cursor
}

func (c cursorWrapper) Read(o Object) error {
	iObj := toInternalObject(o)
	return toExternalError(c.c.Read(iObj))
}

func (c cursorWrapper) Update(o Object) error {
	iObj := toInternalObject(o)
	return toExternalError(c.c.Update(iObj))
}

func (c cursorWrapper) Delete() error {
	return toExternalError(c.c.Delete())
}

func (c cursorWrapper) Next() {
	c.c.Next()
}

func (c cursorWrapper) Valid() bool {
	return c.c.Valid()
}
