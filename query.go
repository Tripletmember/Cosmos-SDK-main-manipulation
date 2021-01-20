package crud

import (
	"fmt"
	"github.com/iov-one/cosmos-sdk-crud/internal/store"
	"github.com/iov-one/cosmos-sdk-crud/internal/store/types"
)

type QueryStatement interface {
	Where() WhereStatement
}

type WhereStatement interface {
	Index(id IndexID) IndexStatement
}

type IndexStatement interface {
	Equals(v []byte) FinalizedIndexStatement
}

type RangeStatement interface {
	Start(start uint64) RangeEndStatement
}

type RangeEndStatement interface {
	End(end uint64) FinalizedIndexStatement
}

type FinalizedIndexStatement interface {
	And() WhereStatement
	WithRange() RangeStatement
	Do() (Cursor, error)
}

func newQuery(s store.Store) *query {
	return &query{
		store:      s,
		andEqualSk: make(map[byte]struct{}),
	}
}

type query struct {
	errs       []error              // errors found during queries
	andEqualSk map[byte]struct{}    // keep track of indexes ID we want to be equal to (or not supported yet)
	sks        []types.SecondaryKey // secondary keys to query for equality
	currSk     types.SecondaryKey   // secondaryKey that is currently being processed
	store      store.Store          // underlying store to use
	start, end uint64               // start and end of query

	consumed bool // used after the query has run Do()
}

func (q *query) And() WhereStatement {
	return q
}

func (q *query) Do() (Cursor, error) {
	// check if there are query errors
	if len(q.errs) != 0 {
		return nil, q.errs[0]
	}
	// check if query has something to query
	if len(q.sks) == 0 {
		return nil, fmt.Errorf("%w: no secondary keys supplied to query", ErrBadArgument)
	}
	// check if query was already run
	if q.consumed {
		return nil, fmt.Errorf("%w: query already consumed", ErrBadArgument)
	}
	// do query
	crs, err := q.store.Query(q.sks, q.start, q.end)
	if err != nil {
		return nil, toExternalError(err)
	}
	// reset query
	q.consumed = true
	// return wrapped cursor
	return cursorWrapper{c: crs}, nil
}

func (q *query) Index(id IndexID) IndexStatement {
	bID := (byte)(id)
	_, ok := q.andEqualSk[bID]
	if ok {
		q.errs = append(q.errs, fmt.Errorf("%w: bad query, equality on index with same id %d", ErrBadArgument, bID))
	}
	q.currSk.ID = bID
	q.andEqualSk[bID] = struct{}{}
	return q
}

func (q *query) Equals(v []byte) FinalizedIndexStatement {
	if v == nil {
		q.errs = append(q.errs, fmt.Errorf("%w: bad query, equality on nil value", ErrBadArgument))
	}
	q.currSk.Value = v
	q.sks = append(q.sks, q.currSk)
	return q
}

func (q *query) WithRange() RangeStatement {
	return q
}

func (q *query) Start(start uint64) RangeEndStatement {
	q.start = start
	return q
}

func (q *query) End(end uint64) FinalizedIndexStatement {
	q.end = end
	return q
}

func (q *query) Where() WhereStatement {
	return q
}
