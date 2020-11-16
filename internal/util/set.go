package util

import "fmt"

func hash(b []byte) string {
	return fmt.Sprintf("%x", b)
}

type ByteSet struct {
	set  map[string]struct{}
	keys map[string][]byte
}

func NewByteSet() ByteSet {
	return ByteSet{
		set:  make(map[string]struct{}),
		keys: make(map[string][]byte),
	}
}

func (s ByteSet) Insert(b []byte) {
	h := hash(b)
	s.set[h] = struct{}{}
	s.keys[h] = b
}

func (s ByteSet) Has(b []byte) (ok bool) {
	_, ok = s.set[hash(b)]
	return ok
}

func (s ByteSet) Len() int {
	return len(s.set)
}

func (s ByteSet) Range() [][]byte {
	r := make([][]byte, len(s.keys))
	i := 0
	for _, v := range s.keys {
		r[i] = v
		i++
	}
	return r
}
