package score

import (
	"bean/internal/trace"
	"context"
)

type Score map[string]float32

type TracesScorer interface {
	Score(ctx context.Context, traces []trace.Trace) (Score, error)
}
