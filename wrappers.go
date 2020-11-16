package crud

import (
	"errors"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/store"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/store/types"
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
		pk:  o.PrimaryKey(),
		sks: sks,
	}
}

type internalObjectWrapper struct {
	pk  []byte
	sks []types.SecondaryKey
}

func (i internalObjectWrapper) SecondaryKeys() []types.SecondaryKey {
	return i.sks
}

func (i internalObjectWrapper) PrimaryKey() []byte {
	return i.pk
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
