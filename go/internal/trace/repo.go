package trace

import (
	"bean/internal/utils"
	"sync"
	"time"
)

type TracesRepository struct {
	length int
	ttl    time.Duration

	traces map[string]*utils.RingBuffer[Trace]
	mu     sync.RWMutex
}

func (tr *TracesRepository) Append(id string, t Trace) {
	tr.mu.RLock()
	buffer, found := tr.traces[id]
	tr.mu.RUnlock()
	if !found {
		tr.mu.Lock()
		if buffer, found = tr.traces[id]; !found {
			buffer = utils.NewRingBuffer[Trace](tr.length)
			tr.traces[id] = buffer
		}
		tr.mu.Unlock()
	}
	buffer.Push(t)
}

func (tr *TracesRepository) Get(id string) ([]Trace, bool) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	buffer, found := tr.traces[id]
	if !found {
		return nil, false
	}
	return buffer.ToSlice(), true
}

func NewTracesRepository(length int, ttl time.Duration) *TracesRepository {
	repo := TracesRepository{
		length: length,
		ttl:    ttl,
		traces: make(map[string]*utils.RingBuffer[Trace]),
	}
	return &repo
}
