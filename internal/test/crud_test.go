package test

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crud "github.com/iov-one/cosmos-sdk-crud"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
)

const starnameDelimiter string = "*"
const starnameOwnerIndex crud.IndexID = 0x1
const starnameDomainIndex crud.IndexID = 0x2

// assert Object is implemented by test objects
var _ = crud.Object(NewTestStarname("", "", ""))

type TestStarname struct {
	*types.TestStarname
}

func NewTestStarname(owner, domain, name string) *TestStarname {
	testObject := types.TestStarname{
		Owner:  owner,
		Domain: domain,
		Name:   &name,
	}
	object := TestStarname{&testObject}

	return &object
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
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("crud.internal.test",
		(*crud.Object)(nil),
		&TestStarname{},
	)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	key := sdk.NewKVStoreKey("crud_test")
	mdb := db.NewMemDB()
	ms := store.NewCommitMultiStore(mdb)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, mdb)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now()}, true, log.NewNopLogger())
	db := ctx.KVStore(key)
	store := crud.NewStore(cdc, db, nil)

	// populate the store and test vectors
	domains := []string{"iov", "cosmos"}
	accounts := []string{"", "binance", "coinbase", "kraken"}
	owners := []string{"antoine", "dave"}
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

	t.Run("success on primary key", func(t *testing.T) {
		// TODO
	})
	t.Run("success on un-ranged owned accounts", func(t *testing.T) {
		for _, owner := range owners {
			cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte(owner)).Do()
			if err != nil {
				t.Fatal(err)
			}
			wants := starnamesByOwner[owner]
			debugStarnames(fmt.Sprintf("%s un-ranged wants", owner), wants)
			for i := 0; cursor.Valid(); cursor.Next() {
				starname := NewTestStarname("", "", "")
				if err := cursor.Read(starname); err != nil {
					t.Fatal(err)
				}
				debugStarname(starname)
				if i > len(wants) {
					t.Fatal("got more than expected")
				}
				if err := wants[i].Equals(starname); err != nil {
					t.Fatal(errors.Wrapf(err, "byOwner[%s][%d]: %s != %s", owner, i, wants[i].GetStarname(), starname.GetStarname()))
				}
			}
		}
	})
	t.Run("success on ranged owned accounts", func(t *testing.T) {
		for _, owner := range owners {
			continue // TODO: DELETEME
			owned := starnamesByOwner[owner]
			end := len(owned)
			for i := 0; i < end; i++ {
				cursor, err := store.Query().Where().Index(starnameOwnerIndex).Equals([]byte(owner)).WithRange().Start(uint64(i)).End(uint64(end)).Do()
				if err != nil {
					t.Fatal(err)
				}
				n := 0
				for ; cursor.Valid(); cursor.Next() {
					starname := NewTestStarname("", "", "")
					if err := cursor.Read(starname); err != nil {
						t.Fatal(err)
					}
					debugStarname(starname)
					n++
				}
				if n != end-i {
					t.Fatalf("expected %d but got %d", end-i, n)
				}
			}
		}
	})
}

func debugStarname(starname *TestStarname) {
	if len(os.Args) > 0 {
		fmt.Printf("%16s %-32x %v %v\n", starname.GetStarname(), starname.PrimaryKey(), starname.SecondaryKeys()[0], starname.SecondaryKeys()[1])
	}
}

func debugStarnames(name string, starnames []*TestStarname) {
	if len(os.Args) > 0 {
		fmt.Printf("___  %s ___\n", name)
		for _, starname := range starnames {
			debugStarname(starname)
		}
		fmt.Printf("___ ~%s ___\n", name)
	}
}
