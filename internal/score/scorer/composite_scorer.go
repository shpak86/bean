package scorer

import (
	"bean/internal/score"
	"bean/internal/trace"
	"context"
	"errors"
)

// CompositeScorer is a composite scorer implementation that aggregates scores
// from multiple nested scorers. To compute the final score, it retrieves
// behavioral traces by session ID from the repository and passes them
// to each scorer. The resulting scores are summed by key and normalized
// to the range [0.0, 1.0].
//
// CompositeScorer is thread-safe, provided that all nested scorers and
// the trace repository (tracesRepo) are also thread-safe.
type CompositeScorer struct {
	scorers    []score.TracesScorer    // list of scorers whose scores will be combined
	tracesRepo *trace.TracesRepository // repository for retrieving trace data by ID
	ctx        context.Context         // context passed to scorers during evaluation
}

// Score calculates the final score for the given session ID.
// Algorithm:
//  1. Retrieves the list of traces from the repository by the given ID.
//  2. If no traces are found, returns an error.
//  3. Invokes the Score method on each scorer with the context and traces.
//  4. Aggregates all resulting scores by key (summation).
//  5. Clamps each score component to the range [0.0, 1.0].
//
// If any scorer returns an error, execution stops immediately and the error is returned.
//
// Parameters:
//   - id: the session ID used to retrieve traces.
//
// Returns:
//   - score.Score: the final aggregated score.
//   - error: an error if the session is not found or any scorer fails.
func (cs *CompositeScorer) Score(id string) (score.Score, error) {
	result := make(score.Score)
	traces, exists := cs.tracesRepo.Get(id)
	if !exists {
		return result, errors.New("trace id not found: " + id)
	}
	for _, s := range cs.scorers {
		score, err := s.Score(cs.ctx, traces)
		if err != nil {
			return result, err
		}
		for k, v := range score {
			result[k] += v
			if result[k] > 1.0 {
				result[k] = 1.0
			} else if result[k] < 0.0 {
				result[k] = 0.0
			}
		}
	}
	return result, nil
}

// NewCompositeScorer creates a new instance of CompositeScorer.
//
// Parameters:
//   - scorers: a list of scorers to be used for score calculation.
//   - tracesRepo: the trace repository from which trace data will be loaded by session ID.
//
// Returns a pointer to the newly created CompositeScorer instance.
// The context is initialized to context.Background() by default.
func NewCompositeScorer(scorers []score.TracesScorer, tracesRepo *trace.TracesRepository) *CompositeScorer {
	return &CompositeScorer{ctx: context.Background(), scorers: scorers, tracesRepo: tracesRepo}
}
