package scorer

import (
	"bean/internal/score"
	"bean/internal/score/rule"
	"bean/internal/trace"
	"context"
	"log/slog"
)

type RulesScorer struct {
	// rules — list of rules applied when calculating the score.
	// Rules are processed in declaration order; each can contribute to the final score.
	rules    []rule.Rule
	min, max float32
}

func (rs *RulesScorer) Score(context context.Context, traces []trace.Trace) (score.Score, error) {
	score := make(score.Score)

	for _, trace := range traces {
		for _, rule := range rs.rules {
			delta, err := rule.Eval(trace)
			if err != nil {
				slog.Error("rule eval", "error", err, "rule", rule, "trace", trace)
				continue
			}

			for key, d := range delta {
				newScore := score[key] + d
				switch {
				case newScore < rs.min:
					score[key] = rs.min
				case newScore > rs.max:
					score[key] = rs.max
				default:
					score[key] = newScore
				}
			}
		}
	}

	return score, nil
}

func NewRulesScorer(rules []rule.Rule, min, max float32) *RulesScorer {
	scorer := RulesScorer{
		rules: rules,
		min:   min,
		max:   max,
	}
	return &scorer
}
