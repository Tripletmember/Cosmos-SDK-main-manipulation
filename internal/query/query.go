package query

import (
	"fmt"

	crud "github.com/iov-one/cosmos-sdk-crud"
)

type StoreWithDirectQuery interface {
	crud.Store
	DoDirectQuery(sks []crud.SecondaryKey, start, end uint64) (crud.Cursor, error)
}

func NewQuery(s StoreWithDirectQuery) *query {
	return &query{
		store:      s,
		andEqualSk: make(map[byte]struct{}),
	}
}

type query struct {
	errs       []error              // errors found during queries
	andEqualSk map[byte]struct{}    // keep track of indexes ID we want to be equal to (or not supported yet)
	sks        []crud.SecondaryKey  // secondary keys to query for equality
	currSk     crud.SecondaryKey    // secondaryKey that is currently being processed
	store      StoreWithDirectQuery // underlying store to use
	start, end uint64               // start and end of query

	consumed bool // used after the query has run Do()
}

func (q *query) And() crud.WhereStatement {
	return q
}

func (q *query) Do() (crud.Cursor, error) {
	// check if there are query errors
	if len(q.errs) != 0 {
		return nil, q.errs[0]
	}
	// check if query was already run
	if q.consumed {
		return nil, fmt.Errorf("%w: query already consumed", crud.ErrBadArgument)
	}
	// do query
	crs, err := q.store.DoDirectQuery(q.sks, q.start, q.end)
	if err != nil {
		return nil, err
	}
	// reset query
	q.consumed = true
	// return wrapped cursor
	return crs, nil
}

func (q *query) Index(id crud.IndexID) crud.IndexStatement {
	bID := (byte)(id)
	_, ok := q.andEqualSk[bID]
	if ok {
		q.errs = append(q.errs, fmt.Errorf("%w: bad query, equality on index with same id %d", crud.ErrBadArgument, bID))
	}
	q.currSk.ID = crud.IndexID(bID)
	q.andEqualSk[bID] = struct{}{}
	return q
}

func (q *query) Equals(v []byte) crud.FinalizedIndexStatement {
	if v == nil {
		q.errs = append(q.errs, fmt.Errorf("%w: bad query, equality on nil value", crud.ErrBadArgument))
	}
	q.currSk.Value = v
	q.sks = append(q.sks, q.currSk)
	return q
}

func (q *query) WithRange() crud.RangeStatement {
	return q
}

func (q *query) Start(start uint64) crud.RangeEndStatement {
	q.start = start
	return q
}

func (q *query) End(end uint64) crud.FinalizedIndexStatement {
	q.end = end
	return q
}

func (q *query) Where() crud.WhereStatement {
	return q
}
