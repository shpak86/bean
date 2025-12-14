package trace

import (
	"bean/internal/utils"
	"sync"
	"time"
)

// TracesRepository — a thread-safe storage for traces with automatic cleanup of outdated records.
// For each identifier (id), a fixed-size ring buffer is maintained.
// Traces that have not been updated longer than the specified TTL are deleted by a background process.
//
// Example usage:
//
// repo := trace.NewTracesRepository(10, 5*time.Minute)
// go repo.Serve() // start background cleanup
// repo.Append("user-123", trace.Trace{"MouseMoves": 5})
type TracesRepository struct {
	length        int                                  // maximum number of traces per identifier
	ttl           time.Duration                        // trace lifetime; after this it is considered outdated
	traces        map[string]*utils.RingBuffer[Trace] // trace storage by ID
	tracesUpdates map[string]time.Time                // last update time for each ID
	cleanTicker   *time.Ticker                         // ticker for periodic cleanup
	tracesMu      sync.RWMutex                         // mutex to protect access to maps
}

// Append adds trace t to the buffer associated with the specified identifier id.
// If there is no buffer for the given id, it is created automatically.
// The last update time for id is updated when creating or on first addition.
// The method is thread-safe.
func (tr *TracesRepository) Append(id string, t Trace) {
	tr.tracesMu.RLock()
	buffer, found := tr.traces[id]
	tr.tracesMu.RUnlock()

	if !found {
		tr.tracesMu.Lock()
		// Double-checked locking
		if buffer, found = tr.traces[id]; !found {
			buffer = utils.NewRingBuffer[Trace](tr.length)
			tr.traces[id] = buffer
			// Update last update time
			tr.tracesUpdates[id] = time.Now()
		}
		tr.tracesMu.Unlock()
	}

	buffer.Push(t)
}

// Get returns a copy of all traces for the specified identifier id in order from old to new.
// If traces for the given id are missing, returns (nil, false).
// The method is thread-safe.
func (tr *TracesRepository) Get(id string) ([]Trace, bool) {
	tr.tracesMu.Lock()
	defer tr.tracesMu.Unlock()

	buffer, found := tr.traces[id]
	if !found {
		return nil, false
	}

	return buffer.ToSlice(), true
}

// Serve starts a background goroutine that periodically (once a minute) checks
// and removes outdated traces — those where more than ttl has passed since the last update.
// The method blocks execution and should be called in a separate goroutine:
//
// go repo.Serve()
//
// Use the Stop method to stop.
func (tr *TracesRepository) Serve() {
	tr.cleanTicker = time.NewTicker(time.Minute)
	for range tr.cleanTicker.C {
		var outdated []string

		// Collect list of outdated IDs under read lock
		tr.tracesMu.RLock()
		now := time.Now()
		for id, ts := range tr.tracesUpdates {
			if now.Sub(ts) > tr.ttl {
				outdated = append(outdated, id)
			}
		}
		tr.tracesMu.RUnlock()

		// Delete outdated records under write lock
		if len(outdated) > 0 {
			tr.tracesMu.Lock()
			for _, id := range outdated {
				delete(tr.traces, id)
				delete(tr.tracesUpdates, id)
			}
			tr.tracesMu.Unlock()
		}
	}
}

// Stop stops background cleanup by stopping the ticker.
// Should be called on shutdown to prevent resource leaks.
// The method is safe to call even if Serve has not been started yet.
func (tr *TracesRepository) Stop() {
	if tr.cleanTicker != nil {
		tr.cleanTicker.Stop()
	}
}

// NewTracesRepository creates a new instance of trace storage.
// Parameters:
// - length: maximum number of traces stored per identifier (buffer rewrites in a circle).
// - ttl: time after which inactive traces are considered outdated and removed by the background process.
//
// Returns a pointer to a new TracesRepository instance.
// To start automatic cleanup, call Serve in a separate goroutine.
func NewTracesRepository(length int, ttl time.Duration) *TracesRepository {
	repo := TracesRepository{
		length:        length,
		ttl:           ttl,
		traces:        make(map[string]*utils.RingBuffer[Trace]),
		tracesUpdates: make(map[string]time.Time),
	}

	return &repo
}
