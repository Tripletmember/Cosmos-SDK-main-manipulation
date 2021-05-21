package iterator

type KeyIterator struct {
	isValid bool
	value   []byte

	nextValue func() ([]byte, bool)
}

func NewKeyIterator(next func() ([]byte, bool)) *KeyIterator {
	// Move to first element
	value, valid := next()
	return &KeyIterator{
		isValid:   valid,
		value:     value,
		nextValue: next,
	}
}

func (it *KeyIterator) Next() {
	it.value, it.isValid = it.nextValue()
}

func (it *KeyIterator) Valid() bool {
	return it.isValid
}

func (it *KeyIterator) Get() []byte {
	return it.value
}

func (it *KeyIterator) Collect() [][]byte {
	data := make([][]byte, 0)
	for ; it.Valid(); it.Next() {
		data = append(data, it.Get())
	}
	return data
}

type NilIterator struct{}

func (it NilIterator) Next()             {}
func (it NilIterator) Valid() bool       { return false }
func (it NilIterator) Get() []byte       { return nil }
func (it NilIterator) Collect() [][]byte { return make([][]byte, 0) }
