package utils

import (
	"testing"
)

func TestRingBuffer_NewRingBuffer(t *testing.T) {
	t.Run("positive size", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		if rb == nil {
			t.Fatal("expected non-nil buffer")
		}

		if rb.Cap() != 3 {
			t.Errorf("expected cap=3, got %d", rb.Cap())
		}

		if rb.Len() != 0 {
			t.Errorf("expected len=0, got %d", rb.Len())
		}
	})

	t.Run("zero size panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for size=0")
			}
		}()
		NewRingBuffer[int](0)
	})

	t.Run("negative size panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for size<0")
			}
		}()
		NewRingBuffer[int](-1)
	})
}

func TestRingBuffer_Push(t *testing.T) {
	rb := NewRingBuffer[int](3)

	rb.Push(1)
	if rb.Len() != 1 {
		t.Errorf("expected len=1, got %d", rb.Len())
	}

	rb.Push(2)
	rb.Push(3)
	if rb.Len() != 3 {
		t.Errorf("expected len=3 after 3 pushes, got %d", rb.Len())
	}

	// Check contents
	expected := []int{1, 2, 3}
	for i, exp := range expected {
		if got := rb.At(i); got != exp {
			t.Errorf("At(%d): expected %d, got %d", i, exp, got)
		}
	}
}

func TestRingBuffer_OverwriteOnFull(t *testing.T) {
	rb := NewRingBuffer[int](3)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	rb.Push(4) // should displace 1

	expected := []int{2, 3, 4}

	if rb.Len() != 3 {
		t.Errorf("len should still be 3, got %d", rb.Len())
	}

	for i, exp := range expected {
		if got := rb.At(i); got != exp {
			t.Errorf("At(%d): expected %d, got %d", i, exp, got)
		}
	}
}

func TestRingBuffer_ContinuousOverwrite(t *testing.T) {
	rb := NewRingBuffer[string](2)

	rb.Push("a")
	rb.Push("b")
	rb.Push("c") // displaces "a"
	rb.Push("d") // displaces "b"

	expected := []string{"c", "d"}
	slice := rb.ToSlice()

	for i, exp := range expected {
		if slice[i] != exp {
			t.Errorf("ToSlice[%d]: expected %s, got %s", i, exp, slice[i])
		}
	}
}

func TestRingBuffer_At_IndexOutOfBounds(t *testing.T) {
	rb := NewRingBuffer[int](3)
	rb.Push(10)

	t.Run("negative index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on negative index")
			}
		}()
		_ = rb.At(-1)
	})

	t.Run("index >= len", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on index >= len")
			}
		}()
		_ = rb.At(1) // len=1 → index=1 out of bounds
	})
}

func TestRingBuffer_ToSlice(t *testing.T) {
	rb := NewRingBuffer[int](4)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	slice := rb.ToSlice()
	expected := []int{1, 2, 3}

	if len(slice) != len(expected) {
		t.Fatalf("slice len: expected %d, got %d", len(expected), len(slice))
	}

	for i, exp := range expected {
		if slice[i] != exp {
			t.Errorf("slice[%d]: expected %d, got %d", i, exp, slice[i])
		}
	}
}

func TestRingBuffer_FullOverwriteSequence(t *testing.T) {
	rb := NewRingBuffer[int](3)

	for i := 1; i <= 6; i++ {
		rb.Push(i)
	}

	// After 6 pushes to a buffer of size 3: should have [4,5,6]
	expected := []int{4, 5, 6}
	slice := rb.ToSlice()

	for i, exp := range expected {
		if slice[i] != exp {
			t.Errorf("after full overwrite: expected %d at %d, got %d", exp, i, slice[i])
		}
	}
}

func TestRingBuffer_CapAndLen(t *testing.T) {
	rb := NewRingBuffer[struct{}](5)

	for i := 0; i < 7; i++ {
		rb.Push(struct{}{})
		if rb.Len() > rb.Cap() {
			t.Errorf("len (%d) > cap (%d) after push %d", rb.Len(), rb.Cap(), i+1)
		}
	}

	if rb.Cap() != 5 {
		t.Errorf("cap changed: expected 5, got %d", rb.Cap())
	}
}
