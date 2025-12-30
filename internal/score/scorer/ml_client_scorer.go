package scorer

import (
	"bean/internal/score"
	"bean/internal/trace"
	"context"
	"net/url"
)

type MLScorerClient struct {
}

func (rs *MLScorerClient) Score(context context.Context, traces []trace.Trace) (score.Score, error) {
	return make(score.Score), nil
}

func NewMLScorerClient(url url.URL, model string) *MLScorerClient {
	return &MLScorerClient{}
}
