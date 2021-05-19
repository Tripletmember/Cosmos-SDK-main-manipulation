package util

import (
	"errors"
	"testing"
)

func TestRange(t *testing.T) {
	r, err := NewRange(3, 10)
	checkErr(t, err)

	t.Run("success", func(t *testing.T) {
		stopIter, inRange := false, false
		i, n := 0, 0
		for !stopIter {
			inRange, stopIter = r.CheckAndMoveForward()
			if inRange {
				i++
				if i == 1 && n != 3 {
					t.Log("Range started at the wrong index (expecting 3, actual", n, ")")
					t.Fail()
				}
			}
			if !stopIter {
				n++
			}
		}

		if i != 7 || n != 10 {
			t.Fatal("Range ended at the wrong index")
		}

	})

	t.Run("reusing range", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			inRange, stopIter := r.CheckAndMoveForward()
			if inRange || !stopIter {
				t.Fatalf("An already consumed range should always return false, true but returned %v, %v", inRange, stopIter)
			}
		}
	})

	t.Run("infinite ranges", func(t *testing.T) {
		const aSufficientlyLargeNumber = 5000

		r, err := NewRange(0, 0)
		checkErr(t, err)

		for i := 0; i < aSufficientlyLargeNumber; i++ {
			inRange, stopIter := r.CheckAndMoveForward()
			if !inRange || stopIter {
				t.Fatal("An infinite range should not end")
			}
		}

		r, err = NewRange(3, 0)
		checkErr(t, err)

		for i := 0; i < 3; i++ {
			inRange, _ := r.CheckAndMoveForward()
			if inRange {
				t.Fatal("Range should not have begun")
			}
		}

		for i := 3; i < aSufficientlyLargeNumber; i++ {
			inRange, stopIter := r.CheckAndMoveForward()
			if !inRange || stopIter {
				t.Fatal("An infinite range should not end")
			}
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		_, err := NewRange(5, 3)
		if !errors.Is(err, errBadRange) {
			t.Fatal("Unexpected error : ", err)
		}
	})

	t.Run("empty range", func(t *testing.T) {
		_, err := NewRange(1, 1)
		if !errors.Is(err, errBadRange) {
			t.Fatal("Unexpected error : ", err)
		}

	})

	t.Run("single element range", func(t *testing.T) {
		r, err := NewRange(0, 1)
		checkErr(t, err)

		inRange, stop := r.CheckAndMoveForward()
		if !inRange || stop {
			t.Fatal("This range should not be empty")
		}

		inRange, stop = r.CheckAndMoveForward()
		if inRange || !stop {
			t.Fatal("This range should be over")
		}
	})

}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Failed with fatal error : ", err)
	}
	return
}
