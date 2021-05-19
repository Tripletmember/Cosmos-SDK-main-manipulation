package util

import (
	"reflect"
	"testing"
)

func TestSet(t *testing.T) {
	s := NewByteSet()
	o1 := []byte("abc")
	o2 := []byte("def")
	o3 := []byte("abcd")

	t.Run("insert/different", func(t *testing.T) {

		s.Insert(o1)
		s.Insert(o2)
		s.Insert(o3)

		if s.Len() != 3 {
			t.Fatal("Expected set length 3, got", s.Len())
		}

		if !s.Has(o1) || !s.Has(o2) || !s.Has(o3) {
			t.Fatal("Missing values in set")
		}

	})
	t.Run("insert/duplicates", func(t *testing.T) {
		o4 := []byte("abc")
		s.Insert(o4)
		o5 := []byte("def")
		s.Insert(o5)

		if s.Len() != 3 {
			t.Fatal("The set inserted duplicates : expected length 3, got", s.Len())
		}

		if !s.Has(o1) || !s.Has(o2) || !s.Has(o3) {
			t.Fatal("Missing values in set")
		}
	})
	t.Run("insert/empty", func(t *testing.T) {
		// Ensure this does not fail, should only insert one element
		s.Insert(nil)
		s.Insert([]byte{})

		if s.Len() != 4 {
			t.Fatal("The set does not inserted an empty byte array : expected length 4, got", s.Len())
		}

		if !s.Has([]byte{}) {
			t.Fatal("Missing values in set")
		}
	})
	t.Run("range/empty", func(t *testing.T) {
		s := NewByteSet()
		array := s.Range()

		if array == nil || len(array) != 0 {
			t.Fatal("Invalid set range when the set is empty")
		}
	})
	t.Run("range/order", func(t *testing.T) {
		expected := [][]byte{{}, o1, o3, o2}
		actual := s.Range()

		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("Invalid range, expecting %v, got %v", expected, actual)
		}
	})

}
