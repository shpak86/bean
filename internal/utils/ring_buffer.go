package utils

import "sync"

// RingBuffer represents a fixed-size circular buffer that stores elements of type T.
// When adding new elements to a full buffer, the oldest element is automatically replaced.
// Elements are stored in order of arrival: from oldest to newest.
//
// Example usage:
//
// rb := NewRingBuffer[int](3)
// rb.Push(1)
// rb.Push(2)
// rb.Push(3)
// rb.Push(4) // element 1 will be displaced
// fmt.Println(rb.ToSlice()) // [2 3 4]
type RingBuffer[T any] struct {
	data  []T       // internal array for storing elements
	size  int       // buffer capacity (maximum number of elements)
	count int       // current number of elements in the buffer
	head  int       // index of the oldest element (where reading starts)
	tail  int       // index of the next free position for writing
	mu    sync.RWMutex
}

// NewRingBuffer creates a new ring buffer of the specified size.
// The size parameter must be a positive number, otherwise the call will panic.
//
// Example:
//
// rb := NewRingBuffer[string](5)
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	if size <= 0 {
		panic("ring buffer size must be positive")
	}

	return &RingBuffer[T]{
		data: make([]T, size),
		size: size,
	}
}

// Push adds an element to the end of the buffer.
// If the buffer is full, the oldest element will be automatically replaced.
//
// Parameter item — the element to add.
func (rb *RingBuffer[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	} else {
		// Buffer is full — shift the beginning (displace the oldest element)
		rb.head = (rb.head + 1) % rb.size
	}
}

// Len returns the current number of elements in the buffer.
// Value is always in the range [0, Cap()].
func (rb *RingBuffer[T]) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// Cap returns the maximum capacity of the buffer — the number of elements
// it can store without displacement.
func (rb *RingBuffer[T]) Cap() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// At returns the element at index i, where i=0 is the oldest element, i=Len()-1 is the newest.
// If the index is out of the valid range [0, Len()), it panics.
//
// Parameter i — element index (from 0 to Len()-1).
func (rb *RingBuffer[T]) At(i int) T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if i < 0 || i >= rb.count {
		panic("index out of range")
	}

	return rb.data[(rb.head+i)%rb.size]
}

// ToSlice returns a copy of all buffer elements as a slice.
// Elements follow in order from oldest to newest.
//
// Result:
// - A new slice of length Len() containing all current elements.
// - If the buffer is empty, returns an empty slice.
func (rb *RingBuffer[T]) ToSlice() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]T, rb.count)
	for i := 0; i < rb.count; i++ {
		result[i] = rb.data[(rb.head+i)%rb.size]
	}

	return result
}
