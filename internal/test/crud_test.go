package test

import (
	"fmt"
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
	domains := []string{"iov", "cosmos"}
	accounts := []string{"", "binance", "coinbase", "kraken"}
	owners := []string{"antoine", "dave"}
	// KVStore insertion order matters, so we have to use a map for
	// future-proofing on change of starname data when dealing with
	// the test vectors.
	starnames := make(map[string]*TestStarname, 0)
	starnamesByOwner := make(map[string]map[string]*TestStarname)

	n := len(owners)
	for i, domain := range domains {
		for j, account := range accounts {
			owner := owners[(i+j)%n] // pseudo random owner
			created := NewTestStarname(owner, domain, account)
			if err := store.Create(created); err != nil {
				t.Fatal(err)
			}
			starname := created.GetStarname()
			starnames[starname] = created
			if _, ok := starnamesByOwner[owner]; !ok {
				starnamesByOwner[owner] = make(map[string]*TestStarname, 0)
			}
			starnamesByOwner[owner][starname] = created
		}
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
			for i := 0; cursor.Valid(); cursor.Next() {
				starname := NewTestStarname("", "", "")
				if err := cursor.Read(starname); err != nil {
					t.Fatal(err)
				}
				//fmt.Printf("%16s %-32x %v %v\n", starname.GetStarname(), starname.PrimaryKey(), starname.SecondaryKeys()[0], starname.SecondaryKeys()[1])
				if i > len(wants) {
					t.Fatal("got more than expected")
				}
				if err := wants[starname.GetStarname()].Equals(starname); err != nil {
					t.Fatal(errors.Wrapf(err, "byOwner[%s][%d]: %s != %s", owner, i, wants[starname.GetStarname()].GetStarname(), starname.GetStarname()))
				}
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
				for ; cursor.Valid(); cursor.Next() {
					starname := NewTestStarname("", "", "")
					if err := cursor.Read(starname); err != nil {
						t.Fatal(err)
					}
					//fmt.Printf("%16s %-32x %v %v\n", starname.GetStarname(), starname.PrimaryKey(), starname.SecondaryKeys()[0], starname.SecondaryKeys()[1])
					n++
				}
				if n != end-i {
					t.Fatalf("expected %d but got %d", end-i, n)
				}
			}
		}
	})
}
