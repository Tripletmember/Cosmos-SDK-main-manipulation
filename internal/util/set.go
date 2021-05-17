package util

func hash(b []byte) string {
	return string(b)
}

type ByteSet struct {
	keys map[string][]byte
}

func NewByteSet() ByteSet {
	return ByteSet{
		keys: make(map[string][]byte),
	}
}

func (s ByteSet) Insert(b []byte) {
	h := hash(b)
	s.keys[h] = b
}

func (s ByteSet) Has(b []byte) (ok bool) {
	_, ok = s.keys[hash(b)]
	return ok
}

func (s ByteSet) Len() int {
	return len(s.keys)
}

func (s ByteSet) Range() [][]byte {
	r := make([][]byte, len(s.keys))
	i := 0
	for _, v := range s.keys {
		r[i] = v
		i++
	}
	SortByteSlice(r) // deal with go's non-deterministic range on maps
	return r
}
