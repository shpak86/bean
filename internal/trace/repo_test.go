package trace

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewTracesRepository проверяет, что репозиторий создаётся с правильными параметрами
func TestNewTracesRepository(t *testing.T) {
	length := 5
	ttl := 10 * time.Minute

	repo := NewTracesRepository(length, ttl)

	assert.Equal(t, length, repo.length, "length should match")
	assert.Equal(t, ttl, repo.ttl, "ttl should match")
	assert.NotNil(t, repo.traces, "traces map should be initialized")
	assert.Empty(t, repo.traces, "traces map should be empty initially")
}

// TestTracesRepository_Append проверяет добавление трейсов в буфер по ID
func TestTracesRepository_Append(t *testing.T) {
	repo := NewTracesRepository(2, 0)

	trace1 := Trace{"MouseMoves": 10, "Clicks": 2}
	trace2 := Trace{"MouseMoves": 5, "Clicks": 1}
	trace3 := Trace{"MouseMoves": 7, "Clicks": 3}

	// Добавляем два трейса — должно поместиться
	repo.Append("user1", trace1)
	repo.Append("user1", trace2)

	// Проверяем, что оба добавились
	traces, ok := repo.Get("user1")
	assert.True(t, ok, "expected traces for user1 to exist")
	assert.Len(t, traces, 2)
	assert.Equal(t, trace1, traces[0], "first trace should match")
	assert.Equal(t, trace2, traces[1], "second trace should match")

	// Добавляем третий — должен вытеснить первый
	repo.Append("user1", trace3)

	traces, _ = repo.Get("user1")
	assert.Len(t, traces, 2)
	assert.Equal(t, trace2, traces[0], "after overwrite, first should be trace2")
	assert.Equal(t, trace3, traces[1], "after overwrite, second should be trace3")
}

// TestTracesRepository_Get проверяет получение трейсов по ID
func TestTracesRepository_Get(t *testing.T) {
	repo := NewTracesRepository(3, 0)

	trace1 := Trace{"MouseMoves": 1, "Clicks": 1}
	trace2 := Trace{"MouseMoves": 2, "Clicks": 2}

	repo.Append("user1", trace1)
	repo.Append("user1", trace2)

	// Проверка существующего ID
	traces, ok := repo.Get("user1")
	assert.True(t, ok, "expected Get to return true for existing ID")
	assert.Len(t, traces, 2)
	assert.Equal(t, []Trace{trace1, trace2}, traces, "retrieved traces should match expected")

	// Проверка несуществующего ID
	_, ok = repo.Get("user2")
	assert.False(t, ok, "expected Get to return false for non-existent ID")
}

// TestTracesRepository_ConcurrentAppend проверяет потокобезопасность Append
func TestTracesRepository_ConcurrentAppend(t *testing.T) {
	repo := NewTracesRepository(100, 0)
	iterations := 1000

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				trace := Trace{
					"MouseMoves": int32(j),
					"Clicks":     int32(j),
				}
				repo.Append(id, trace)
			}
		}(string(rune('A' + i)))
	}

	wg.Wait()

	// Проверим, что все ID создали свои буферы
	for i := 0; i < 10; i++ {
		id := string(rune('A' + i))
		traces, ok := repo.Get(id)
		assert.True(t, ok, "expected traces for ID %s to exist", id)
		assert.NotEmpty(t, traces, "expected non-empty traces for ID %s", id)
		// Последний добавленный элемент должен быть с MouseMoves = iterations-1
		last := traces[len(traces)-1]
		assert.Equal(t, int32(iterations-1), last["MouseMoves"], "last MouseMoves should match for ID %s", id)
	}
}

// TestTracesRepository_RepeatedAppend проверяет, что один и тот же ID использует один и тот же буфер
func TestTracesRepository_RepeatedAppend(t *testing.T) {
	repo := NewTracesRepository(3, 0)

	trace1 := Trace{"MouseMoves": 1}
	trace2 := Trace{"MouseMoves": 2}
	trace3 := Trace{"MouseMoves": 3}
	trace4 := Trace{"MouseMoves": 4}

	repo.Append("user1", trace1)
	repo.Append("user1", trace2)
	repo.Append("user1", trace3)
	repo.Append("user1", trace4) // должен вытеснить trace1

	traces, ok := repo.Get("user1")
	assert.True(t, ok, "expected traces for user1")
	assert.Len(t, traces, 3)

	expected := []Trace{trace2, trace3, trace4}
	assert.Equal(t, expected, traces, "traces should match expected sequence")
}

// TestTracesRepository_MultipleIDs проверяет независимость буферов для разных ID
func TestTracesRepository_MultipleIDs(t *testing.T) {
	repo := NewTracesRepository(2, 0)

	repo.Append("user1", Trace{"MouseMoves": 1})
	repo.Append("user1", Trace{"MouseMoves": 2})
	repo.Append("user1", Trace{"MouseMoves": 3}) // вытеснит 1

	repo.Append("user2", Trace{"MouseMoves": 10})
	repo.Append("user2", Trace{"MouseMoves": 20})

	traces1, _ := repo.Get("user1")
	traces2, _ := repo.Get("user2")

	assert.Len(t, traces1, 2, "user1 should have 2 traces")
	assert.Len(t, traces2, 2, "user2 should have 2 traces")

	expected1 := []Trace{
		{"MouseMoves": 2},
		{"MouseMoves": 3},
	}
	expected2 := []Trace{
		{"MouseMoves": 10},
		{"MouseMoves": 20},
	}

	assert.Equal(t, expected1, traces1, "user1 traces should match")
	assert.Equal(t, expected2, traces2, "user2 traces should match")
}
