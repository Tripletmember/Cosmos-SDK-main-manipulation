package test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	types2 "github.com/iov-one/cosmos-sdk-crud/types"

	"github.com/lucasjones/reggen"

	"github.com/tendermint/tendermint/libs/rand"
)

type Config struct {
	AccountRegex, DomainRegex, ResourceRegex string
}

func (c Config) String() string {
	return fmt.Sprintf("Configuration:\n\tAccount regex : %v\n\tDomain regex : %v\n\tResource regex : %v\n",
		c.AccountRegex, c.DomainRegex, c.ResourceRegex)
}

func newRandomString(regex string) (string, error) {
	return reggen.Generate(regex, 256)
}

func fetchConfig(url string) (*Config, error) {
	resp, err := http.Post(url, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body := make([]byte, 4096)
	// ErrUnexpectedEOF is expected as we over-allocate our slice
	n, err := io.ReadFull(resp.Body, body)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	// Resize the slice to the correct size
	body = body[:n]

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	data = data["result"].(map[string]interface{})
	data = data["configuration"].(map[string]interface{})

	return &Config{
		AccountRegex:  data["valid_account_name"].(string),
		DomainRegex:   data["valid_domain_name"].(string),
		ResourceRegex: data["valid_resource"].(string),
	}, nil
}

func NewRandomStarname(config *Config) *TestStarname {
	domain, _ := newRandomString(config.DomainRegex)
	account, _ := newRandomString(config.AccountRegex)
	resource, _ := newRandomString(config.ResourceRegex)
	owner := "star" + base64.StdEncoding.EncodeToString(rand.Bytes(30))
	return NewTestStarnameWithResource(owner, domain, account, resource)
}

type testFunc func(*testing.T, types2.Store, []*TestStarname)

const nbObjectsInTheStore = 100000
const nbIterations = 50000
const mainnetConfigEndpoint = "https://lcd-private-iov-mainnet-2.iov.one/configuration/query/configuration"

// We need a constant seed in order to be consistent between runs
const randomSeed = int64(123465789)

// TestFuzz runs multiple public functions of this package with random data inputs
func TestFuzz(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	config, err := fetchConfig(mainnetConfigEndpoint)
	if err != nil {
		panic(err)
	}
	fmt.Println("Fetched config from mainnet : ", config)

	rand.Seed(randomSeed)
	fmt.Println("Fuzzing seed : ", randomSeed)

	s := newStarnameStore()

	// Create nbObjectsInTheStore objects and store them to the store and to the objs slice
	objs := make([]*TestStarname, nbObjectsInTheStore)
	for i := 0; i < nbObjectsInTheStore; i++ {
		objs[i] = NewRandomStarname(config)
		err := s.Create(objs[i])
		if errors.Is(err, types2.ErrAlreadyExists) {
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

func testCursorDelete(t *testing.T, store types2.Store, objects []*TestStarname) {
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

	if err := store.Read(obj.PrimaryKey(), testObj); !errors.Is(err, types2.ErrNotFound) {
		t.Fatalf("Cursor deleted failed")
	}

	if err := store.Create(obj); err != nil {
		t.Fatalf("Error while re-inserting deleted starname with cursor : %v", err)
	}

}

func testCursorUpdate(t *testing.T, store types2.Store, objects []*TestStarname) {
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

	obj.Resource = "U" + obj.Resource[1:]
	err = cursor.Update(obj)
	CheckNoError(t, err)

	testObj = NewTestStarname("", "", "")
	err = store.Read(obj.PrimaryKey(), testObj)
	CheckNoError(t, err)

	if err := testObj.Equals(obj); err != nil {
		t.Fatalf("Cursor update failed, expecting %v, got %v (%v)", obj, testObj, err)
	}
}

func testDelete(t *testing.T, store types2.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	err := store.Delete(obj.PrimaryKey())
	CheckNoError(t, err)

	testObj := NewTestStarname("", "", "")
	if err := store.Read(obj.PrimaryKey(), testObj); !errors.Is(err, types2.ErrNotFound) {
		t.Fatalf("Deleted failed")
	}

	if err := store.Create(obj); err != nil {
		t.Fatalf("Error while re-inserting deleted starname : %v", err)
	}

}

func testUpdate(t *testing.T, store types2.Store, objects []*TestStarname) {
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

func testAndQuery(t *testing.T, store types2.Store, objects []*TestStarname) {
	obj := randomObject(objects)
	cursor, err := store.Query().Where().Index(starnameResourceIndex).Equals([]byte(obj.Resource)).
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
		if result.Resource != obj.Resource {
			t.Fatalf("Got an unexpected result : expecting resource %v, got resource %v", obj.Resource, result.Resource)
		}
	}

}

func testSimpleQuery(t *testing.T, store types2.Store, objects []*TestStarname) {
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
