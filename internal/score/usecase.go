package score

import (
	"bean/internal/trace"
	"context"
)

type Score map[string]float32

type TraceScorer interface {
	Score(string) (Score, error)
}

type Scorer interface {
	Score(ctx context.Context, traces []trace.Trace) (Score, error)
}
