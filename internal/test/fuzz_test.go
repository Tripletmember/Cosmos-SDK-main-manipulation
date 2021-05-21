package test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"

	"github.com/tendermint/tendermint/libs/rand"

	crud "github.com/iov-one/cosmos-sdk-crud"
)

const domainMinLength = 4
const domainMaxLength = 16
const domainAlphabet = "-_abcdefghijklmnopqrstuvwxyz0123456789"

const accountMinLength = 0
const accountMaxLength = 64
const accountAlphabet = "-_\\.abcdefghijklmnopqrstuvwxyz0123456789"

func newRandomString(min, max int, alphabet []byte) string {
	length := rand.Int()%(max-min) + min
	alphabetLength := len(alphabet)
	var builder strings.Builder
	for i := 0; i < length; i++ {
		char := alphabet[rand.Int()%alphabetLength]
		builder.WriteByte(char)
	}
	return builder.String()
}

func NewRandomStarname() *TestStarname {
	domain := newRandomString(domainMinLength, domainMaxLength, []byte(domainAlphabet))
	account := newRandomString(accountMinLength, accountMaxLength, []byte(accountAlphabet))
	owner := "star" + base64.StdEncoding.EncodeToString(rand.Bytes(30))
	return NewTestStarname(owner, domain, account)
}

type testFunc func(*testing.T, crud.Store, []*TestStarname)

const nbObjectsInTheStore = 100000
const nbIterations = 50000

// We need a constant seed in order to be consistent between runs
const randomSeed = int64(123465789)

// TestFuzz runs multiple public functions of this package with random data inputs
func TestFuzz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	rand.Seed(randomSeed)

	fmt.Println("Fuzzing seed : ", randomSeed)

	s := newStarnameStore()

	// Create nbObjectsInTheStore objects and store them to the store and to the objs slice
	objs := make([]*TestStarname, nbObjectsInTheStore)
	for i := 0; i < nbObjectsInTheStore; i++ {
		objs[i] = NewRandomStarname()
		err := s.Create(objs[i])
		if errors.Is(err, types.ErrAlreadyExists) {
			// This is very unlikely, but the test should not fail in that case
			// Skip this object but we sant the same number at the end
			i--
			continue
		} else if err != nil {
			t.Fatal("Error while creating an object")
		}
	}

	// Now apply the tests functions in a random order nbIterations times
	testFunctions := []testFunc{testSimpleQuery, testAndQuery, testUpdate, testDelete, testCursorUpdate, testCursorDelete}
	nbTestFunctions := len(testFunctions)
	for i := 0; i < nbIterations; i++ {
		testFuncIndex := rand.Int() % nbTestFunctions
		testFunctions[testFuncIndex](t, s, objs)
	}

}

func testCursorDelete(t *testing.T, store crud.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte(obj.Owner)).Do()
	CheckNoError(t, err)

	testObj := NewTestStarname("", "", "")
	for ; cursor.Valid(); cursor.Next() {
		err := cursor.Read(testObj)
		CheckNoError(t, err)
		if bytes.Equal(testObj.PrimaryKey(), obj.PrimaryKey()) {
			break
		}
	}
	if !cursor.Valid() {
		t.Fatal("Missing starname object in query")
	}

	err = cursor.Delete()
	CheckNoError(t, err)

	if err := store.Read(obj.PrimaryKey(), testObj); !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("Cursor deleted failed")
	}

	if err := store.Create(obj); err != nil {
		t.Fatalf("Error while re-inserting deleted starname with cursor : %v", err)
	}

}

func testCursorUpdate(t *testing.T, store crud.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte(obj.Owner)).Do()
	CheckNoError(t, err)

	testObj := NewTestStarname("", "", "")
	for ; cursor.Valid(); cursor.Next() {
		err := cursor.Read(testObj)
		CheckNoError(t, err)
		if bytes.Equal(testObj.PrimaryKey(), obj.PrimaryKey()) {
			break
		}
	}
	if !cursor.Valid() {
		t.Fatal("Missing starname object in query")
	}

	obj.Owner = "U" + obj.Owner[1:]
	err = cursor.Update(obj)
	CheckNoError(t, err)

	testObj = NewTestStarname("", "", "")
	err = store.Read(obj.PrimaryKey(), testObj)
	CheckNoError(t, err)

	if err := testObj.Equals(obj); err != nil {
		t.Fatalf("Cursor update failed, expecting %v, got %v (%v)", obj, testObj, err)
	}
}

func testDelete(t *testing.T, store crud.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	err := store.Delete(obj.PrimaryKey())
	CheckNoError(t, err)

	testObj := NewTestStarname("", "", "")
	if err := store.Read(obj.PrimaryKey(), testObj); !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("Deleted failed")
	}

	if err := store.Create(obj); err != nil {
		t.Fatalf("Error while re-inserting deleted starname : %v", err)
	}

}

func testUpdate(t *testing.T, store crud.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	obj.Owner = "S" + obj.Owner[1:]
	err := store.Update(obj)
	CheckNoError(t, err)

	testObj := NewTestStarname("", "", "")
	err = store.Read(obj.PrimaryKey(), testObj)
	CheckNoError(t, err)

	if err := testObj.Equals(obj); err != nil {
		t.Fatalf("Update failed, expecting %v, got %v (%v)", obj, testObj, err)
	}
}

func testAndQuery(t *testing.T, store crud.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	cursor, err := store.Query().Where().Index(starnameDomainIndex).Equals([]byte(obj.Domain)).
		And().Index(starnameOwnerIndex).Equals([]byte(obj.Owner)).Do()
	CheckNoError(t, err)
	if !cursor.Valid() {
		t.Fatal("At least one starname expected, got 0")
	}
	for ; cursor.Valid(); cursor.Next() {
		result := NewTestStarname("", "", "")
		err = cursor.Read(result)
		CheckNoError(t, err)
		if result.Owner != obj.Owner {
			t.Fatalf("Got an unexpected result : expecting owner %v, got owner %v", obj.Owner, result.Owner)
		}
		if result.Domain != obj.Domain {
			t.Fatalf("Got an unexpected result : expecting domain %v, got domain %v", obj.Domain, result.Domain)
		}
	}

}

func testSimpleQuery(t *testing.T, store crud.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	cursor, err := store.Query().Where().Index(starnameDomainIndex).Equals([]byte(obj.Domain)).Do()
	CheckNoError(t, err)
	if !cursor.Valid() {
		t.Fatal("At least one starname expected, got 0")
	}
	for ; cursor.Valid(); cursor.Next() {
		result := NewTestStarname("", "", "")
		err = cursor.Read(result)
		CheckNoError(t, err)
		if result.Domain != obj.Domain {
			t.Fatalf("Got an unexpected result : expecting domain %v, got domain %v", obj.Domain, result.Domain)
		}
	}
}

func randomObject(objects []*TestStarname) *TestStarname {
	i := rand.Int() % len(objects)
	return objects[i]
}
