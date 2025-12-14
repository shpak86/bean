package utils

import "sync"

// RingBuffer представляет собой кольцевой буфер фиксированного размера, который хранит элементы типа T.
// При добавлении новых элементов в заполненный буфер, самый старый элемент автоматически заменяется.
// Элементы хранятся в порядке поступления: от самого старого к самому новому.
//
// Пример использования:
//
//	rb := NewRingBuffer[int](3)
//	rb.Push(1)
//	rb.Push(2)
//	rb.Push(3)
//	rb.Push(4) // элемент 1 будет вытеснен
//	fmt.Println(rb.ToSlice()) // [2 3 4]
type RingBuffer[T any] struct {
	data  []T // внутренний массив для хранения элементов
	size  int // ёмкость буфера (максимальное количество элементов)
	count int // текущее количество элементов в буфере
	head  int // индекс самого старого элемента (с которого начинается чтение)
	tail  int // индекс следующей свободной позиции для записи
	mu    sync.RWMutex
}

// NewRingBuffer создаёт новый кольцевой буфер указанного размера.
// Параметр size должен быть положительным числом, иначе вызов приведёт к панике.
//
// Пример:
//
//	rb := NewRingBuffer[string](5)
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	if size <= 0 {
		panic("ring buffer size must be positive")
	}
	return &RingBuffer[T]{
		data: make([]T, size),
		size: size,
	}
}

// Push добавляет элемент в конец буфера.
// Если буфер заполнен, самый старый элемент будет автоматически заменён.
//
// Параметр item — элемент, который нужно добавить.
func (rb *RingBuffer[T]) Push(item T) {
	rb.data[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	} else {
		// Буфер полон — сдвигаем начало (вытесняем самый старый элемент)
		rb.head = (rb.head + 1) % rb.size
	}
}

// Len возвращает текущее количество элементов в буфере.
// Значение всегда находится в диапазоне [0, Cap()].
func (rb *RingBuffer[T]) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// Cap возвращает максимальную ёмкость буфера — количество элементов,
// которое он может хранить без вытеснения.
func (rb *RingBuffer[T]) Cap() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// At возвращает элемент по индексу i, где i=0 — самый старый элемент, i=Len()-1 — самый новый.
// Если индекс выходит за пределы допустимого диапазона [0, Len()), вызывает панику.
//
// Параметр i — индекс элемента (от 0 до Len()-1).
func (rb *RingBuffer[T]) At(i int) T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if i < 0 || i >= rb.count {
		panic("index out of range")
	}
	return rb.data[(rb.head+i)%rb.size]
}

// ToSlice возвращает копию всех элементов буфера в виде слайса.
// Элементы следуют в порядке от самого старого к самому новому.
//
// Результат:
//   - Новый слайс длиной Len(), содержащий все текущие элементы.
//   - Если буфер пуст, возвращается пустой слайс.
func (rb *RingBuffer[T]) ToSlice() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	result := make([]T, rb.count)
	for i := 0; i < rb.count; i++ {
		result[i] = rb.At(i)
	}
	return result
}
