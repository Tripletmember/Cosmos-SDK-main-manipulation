package crud

import (
	"errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fdymylja/cosmos-sdk-oodb/internal/store"
)

// errors mirroring internal crud's error

// ErrNotFound is returned when an index or an object are not found
var ErrNotFound = errors.New("crud: not found")

// ErrAlreadyExists is returned when an index or an object already exist
var ErrAlreadyExists = errors.New("crud: already exists")

// ErrBadArgument is returned when the provided arguments are invalid
var ErrBadArgument = errors.New("crud: bad argument")

// ErrInternal is returned when the store detects internal error which
// might be related to possible state corruption
var ErrInternal = errors.New("crud: internal error")

type OptionFunc func()

// IndexID uniquely identifies an index
// for example an index ID might be
// the email index, which is represented by
// the unique byte identifier 0x0
type IndexID byte

// SecondaryKey represents a secondary key for an object
type SecondaryKey struct {
	// ID represents the index ID, for example the email's index
	// represented through the 0x0 byte
	ID IndexID
	// Value represents the value of the secondary key
	// example, in, email's index, it might be:
	// []byte("email@example.com")
	Value []byte
}

// Object defines a structure that can be saved in the crud store
type Object interface {
	// PrimaryKey is the unique id that identifies the object
	PrimaryKey() []byte
	// SecondaryKeys is an array containing the secondary keys
	// used to map the object
	SecondaryKeys() []SecondaryKey
}

// Store defines the abstract interface of the crud store
type Store interface {
	// Create creates an object
	// errors if the object already exists
	// if the secondary keys provided are invalid
	// or in case of marshalling error
	Create(o Object) error
	// Read reads to the given object using the primary key
	// 'o' is expected to be a pointer
	// fails in case or error unmarshalling or in case
	// the object does not exist
	Read(primaryKey []byte, o Object) error
	// Update updates the given object
	// fails if there are errors marshalling
	// or if the object does not exist
	Update(o Object) error
	// Delete deletes the object from the crud store
	// given the primary key, fails if the object with
	// primary key provided does not exist
	Delete(primaryKey []byte) error
	// Query allows to use query statements to retrieve objects
	// using their secondary keys
	Query() QueryStatement
}

// Cursor defines an objects iterator returned after a query to the crud.Store
type Cursor interface {
	// Read reads the current object to the provided object interface
	Read(o Object) error
	// Update updates the current object using the provided object interface
	Update(o Object) error
	// Delete deletes the current object
	Delete() error
	// Next moves onto the next primary key
	Next()
	// Valid asserts if the cursor is fully consumed or not
	Valid() bool
}

// NewStore instantiates a new store given the codec, a sdk.KVStore, a prefix
// in which to store the data, and the optional options to customize the store
// behaviour
func NewStore(cdc *codec.Codec, db sdk.KVStore, prefix []byte, opts ...OptionFunc) Store {
	return storeWrapper{s: store.NewStore(cdc, db, prefix, toInternalOptions(opts)...)}
}
