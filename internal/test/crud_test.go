package test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crudstore "github.com/iov-one/cosmos-sdk-crud/internal/store"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	crud "github.com/iov-one/cosmos-sdk-crud/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

const starnameDelimiter string = "*"
const starnameOwnerIndex crud.IndexID = 0x1
const starnameDomainIndex crud.IndexID = 0x2
const starnameResourceIndex crud.IndexID = 0x3

// assert Object is implemented by test objects
var _ = crud.Object(NewTestStarname("", "", ""))

type TestStarname struct {
	*types.TestStarname
}

func newStarnameStore() crud.Store {
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("crud.internal.test",
		(*crud.Object)(nil),
		&TestStarname{},
	)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	key := sdk.NewKVStoreKey("crud_test")
	mdb := tmdb.NewMemDB()
	ms := store.NewCommitMultiStore(mdb)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, mdb)
	if err := ms.LoadLatestVersion(); err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now()}, true, log.NewNopLogger())
	db := ctx.KVStore(key)
	return crudstore.NewStore(cdc, db, nil)
}

func NewTestStarnameWithResource(owner, domain, name, resource string) *TestStarname {
	testObject := types.TestStarname{
		Owner:    owner,
		Domain:   domain,
		Name:     &name,
		Resource: resource,
	}
	object := TestStarname{&testObject}

	return &object
}
func NewTestStarname(owner, domain, name string) *TestStarname {
	return NewTestStarnameWithResource(owner, domain, name, "unused")
}

func (o TestStarname) PrimaryKey() (primaryKey []byte) {
	if len(o.Domain) == 0 || o.Name == nil {
		return nil
	}
	key := strings.Join([]string{o.Domain, *o.Name}, starnameDelimiter)
	return []byte(key)
}

func (o TestStarname) SecondaryKeys() (secondaryKeys []crud.SecondaryKey) {
	var sks []crud.SecondaryKey
	// index by owner
	if len(o.Owner) != 0 {
		ownerIndex := crud.SecondaryKey{
			ID:    starnameOwnerIndex,
			Value: []byte(o.Owner),
		}
		sks = append(sks, ownerIndex)
	}
	// index by domain
	if len(o.Domain) != 0 {
		domainIndex := crud.SecondaryKey{
			ID:    starnameDomainIndex,
			Value: []byte(o.Domain),
		}
		sks = append(sks, domainIndex)
	}

	if len(o.Resource) != 0 {
		sks = append(sks, crud.SecondaryKey{ID: starnameResourceIndex, Value: []byte(o.Resource)})
	}
	return sks
}

func (o *TestStarname) GetStarname() string {
	return fmt.Sprintf("%s%s%s", *o.Name, starnameDelimiter, o.Domain)
}

func (o *TestStarname) Equals(x *TestStarname) error {
	if o.Owner != x.Owner {
		return fmt.Errorf("wanted Owner '%s' but got '%s'", o.Owner, x.Owner)
	}
	if o.Domain != x.Domain {
		return fmt.Errorf("wanted Domain '%s' but got '%s'", o.Domain, x.Domain)
	}
	if o.Name == nil || x.Name == nil {
		return fmt.Errorf("wanted a non-nil Name but got '%v' and '%v'", o.Name, x.Name)
	}
	if *o.Name != *x.Name {
		return fmt.Errorf("wanted Name '%s' but got '%s'", *o.Name, *x.Name)
	}
	return nil
}

func Test_Starname(t *testing.T) {
	// setup dependencies
	store := newStarnameStore()

	// populate the store and test vectors
	// The data is not sorted
	domains := []string{"iov", "cosmos"}
	accounts := []string{"", "coinbase", "kraken", "binance"}
	owners := []string{"dave", "antoine"}
	starnames := make([]*TestStarname, 0)
	starnamesByOwner := make(map[string][]*TestStarname)

	n := len(owners)
	for i, domain := range domains {
		for j, account := range accounts {
			owner := owners[(i+j)%n] // pseudo random owner
			starname := NewTestStarname(owner, domain, account)
			if err := store.Create(starname); err != nil {
				t.Fatal(err)
			}
			starnames = append(starnames, starname)
			starnamesByOwner[owner] = append(starnamesByOwner[owner], starname)
		}
	}

	// sort test vectors on primary key
	sort.Slice(starnames, func(i, j int) bool {
		return bytes.Compare(starnames[i].PrimaryKey(), starnames[j].PrimaryKey()) < 0
	})
	debugStarnames("starnames", starnames)
	for owner, slice := range starnamesByOwner {
		sort.Slice(slice, func(i, j int) bool {
			return bytes.Compare(slice[i].PrimaryKey(), slice[j].PrimaryKey()) < 0
		})
		debugStarnames(owner, slice)
	}
	t.Run("success on empty result", func(t *testing.T) {
		cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte("dave_")).Do()
		if err != nil {
			t.Fatal("Unexpected error :", err)
		}
		if cursor.Valid() {
			t.Fatal("Result found when no result expected")
		}
	})
	t.Run("success on and query", func(t *testing.T) {
		expected := []*TestStarname{
			NewTestStarname(owners[1], domains[1], accounts[0]),
			NewTestStarname(owners[1], domains[1], accounts[2]),
		}
		cursor, err := store.Query().
			Where().Index(starnameOwnerIndex).Equals([]byte(owners[1])).
			And().Index(starnameDomainIndex).Equals([]byte(domains[1])).
			Do()

		if err != nil {
			t.Fatal("Unexpected error :", err)
		}
		i := 0
		for ; cursor.Valid(); cursor.Next() {

			if i == 2 {
				t.Fatal("Too many results for query")
			}

			actual := NewTestStarname("", "", "")
			if err := cursor.Read(actual); err != nil {
				t.Fatal("Unexpected error :", err)
			}

			if actual.Equals(expected[i]) != nil {
				t.Fatalf("Starname mismatch, expected %v, got %v", expected[i], actual)
			}
			i++
		}
		if i != 2 {
			t.Fatalf("Missing values for query : expecting %v, got %v", 2, i)
		}

	})
	t.Run("success on primary key", func(t *testing.T) {
		for _, expected := range starnames {

			actual := NewTestStarname("", "", "")
			if err := store.Read(expected.PrimaryKey(), actual); err != nil {
				t.Fatal("Unexpected error :", err)
			}

			if actual.Equals(expected) != nil {
				t.Fatalf("Starname mismatch, expected %v, got %v", expected, actual)
			}
		}
	})

	t.Run("success on select all", func(t *testing.T) {
		cursor, err := store.Query().Do()
		if err != nil {
			t.Fatal("Unexpected error :", err)
		}

		for i, expected := range starnames {
			if !cursor.Valid() {
				t.Fatal("got less than expected, expected length is", len(starnames), "data ended at index", i)
			}

			actual := NewTestStarname("", "", "")
			if err := cursor.Read(actual); err != nil {
				t.Fatal("Unexpected error :", err)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("Starname mismatch, expected %v, got %v", expected, actual)
			}
			cursor.Next()
		}

		if cursor.Valid() {
			t.Fatal("Got more results than expected")
		}
	})

	t.Run("success on un-ranged owned accounts", func(t *testing.T) {
		for _, owner := range owners {
			cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte(owner)).Do()
			if err != nil {
				t.Fatal(err)
			}
			wants := starnamesByOwner[owner]
			for i := 0; cursor.Valid(); cursor.Next() {
				starname := NewTestStarname("", "", "")
				if err := cursor.Read(starname); err != nil {
					t.Fatal(err)
				}
				if i >= len(wants) {
					t.Fatal("got more than expected")
				}
				if err := wants[i].Equals(starname); err != nil {
					t.Fatal(errors.Wrapf(err, "byOwner[%s][%d]: %s != %s", owner, i, wants[i].GetStarname(), starname.GetStarname()))
				}
				i++
			}
		}
	})
	t.Run("success on ranged owned accounts", func(t *testing.T) {
		for _, owner := range owners {
			owned := starnamesByOwner[owner]
			end := len(owned)
			for i := 0; i < end; i++ {
				cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte(owner)).WithRange().Start(uint64(i)).End(uint64(end)).Do()
				if err != nil {
					t.Fatal(err)
				}
				n := 0
				wants := owned[i:]
				for ; cursor.Valid(); cursor.Next() {
					starname := NewTestStarname("", "", "")
					if err := cursor.Read(starname); err != nil {
						t.Fatal(err)
					}
					if err := wants[n].Equals(starname); err != nil {
						t.Fatalf("For owner %v at index %v : expecting %v but got %v", owner, n, wants[n].GetStarname(), starname.GetStarname())
					}
					n++
				}
				if n != end-i {
					t.Fatalf("expected %d but got %d", end-i, n)
				}
			}
		}
	})
}

func BenchmarkQuerySimple_1000_Objs(b *testing.B) {
	benchmarkSingleQuery(b, 1000)
}
func BenchmarkQuerySimple_10000_Objs(b *testing.B) {
	benchmarkSingleQuery(b, 10000)
}
func BenchmarkQuerySimple_100000_Objs(b *testing.B) {
	benchmarkSingleQuery(b, 100000)
}
func BenchmarkQuerySimple_1000000_Objs(b *testing.B) {
	benchmarkSingleQuery(b, 1000000)
}

func BenchmarkQueryAnded_1000_Objs(b *testing.B) {
	benchmarkAndedQuery(b, 1000)
}
func BenchmarkQueryAnded_10000_Objs(b *testing.B) {
	benchmarkAndedQuery(b, 10000)
}
func BenchmarkQueryAnded_100000_Objs(b *testing.B) {
	benchmarkAndedQuery(b, 100000)
}
func BenchmarkQueryAnded_1000000_Objs(b *testing.B) {
	benchmarkAndedQuery(b, 1000000)
}

func BenchmarkQueryAll_1000_Objs(b *testing.B) {
	benchmarkQueryAll(b, 1000)
}

func BenchmarkQueryAll_10000_Objs(b *testing.B) {
	benchmarkQueryAll(b, 10000)
}
func BenchmarkQueryAll_100000_Objs(b *testing.B) {
	benchmarkQueryAll(b, 100000)
}
func BenchmarkQueryAll_1000000_Objs(b *testing.B) {
	benchmarkQueryAll(b, 1000000)
}

func benchmarkSingleQuery(b *testing.B, nbObjects int) {
	benchmarkQuery(b, nbObjects, func(query crud.QueryStatement) (crud.Cursor, error) {
		return query.Where().Index(starnameDomainIndex).Equals([]byte("domain1")).Do()
	})
}

func benchmarkAndedQuery(b *testing.B, nbObjects int) {
	benchmarkQuery(b, nbObjects, func(query crud.QueryStatement) (crud.Cursor, error) {
		return query.Where().Index(starnameDomainIndex).Equals([]byte("domain1a")).
			And().Index(starnameOwnerIndex).Equals([]byte("star1rkwdtq34kq69j7vta3etxdzrat672x6x4y2duv")).Do()
	})
}

func benchmarkQueryAll(b *testing.B, nbObjects int) {
	benchmarkQuery(b, nbObjects, crud.QueryStatement.Do)
}

var testStores map[int]crud.Store = make(map[int]crud.Store)

// benchmarkQuery benchmarks a simple query in a store with nbObjects objects
// This allows to detect if the query execution time evolves non-linearly with the number of objects in the store
func benchmarkQuery(b *testing.B, nbObjects int, doQuery func(crud.QueryStatement) (crud.Cursor, error)) {
	// Cache created stores
	if _, ok := testStores[nbObjects]; !ok {
		testStores[nbObjects] = starnameStoreWithRandomObjects(nbObjects)
	}
	s := testStores[nbObjects]
	obj := NewTestStarname("", "", "")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		curs, err := doQuery(s.Query())
		if err != nil {
			panic(err)
		}
		for ; curs.Valid(); curs.Next() {
			if err := curs.Read(obj); err != nil {
				panic(err)
			}
		}
	}
}

func starnameStoreWithRandomObjects(nbObjects int) crud.Store {
	s := newStarnameStore()

	for i := 0; i < nbObjects; i++ {
		// 100 names per domain
		generatedDomain := "domain" + fmt.Sprintf("%x", i/100)
		// Repeat names over the objects (but ever with the same domain)
		generatedName := "name" + fmt.Sprintf("%x", i%1000)

		owner := "star" + base64.StdEncoding.EncodeToString(rand.Bytes(30))
		starname := NewTestStarname(owner, generatedDomain, generatedName)
		if err := s.Create(starname); err != nil {
			panic(err)
		}
	}
	return s
}

func debugStarname(starname *TestStarname) {
	fmt.Printf("%16s %-32x %v %v\n", starname.GetStarname(), starname.PrimaryKey(), starname.SecondaryKeys()[0], starname.SecondaryKeys()[1])
}

func debugStarnames(name string, starnames []*TestStarname) {
	fmt.Printf("___  %s ___\n", name)
	for _, starname := range starnames {
		debugStarname(starname)
	}
	fmt.Printf("___ ~%s ___\n", name)

}
