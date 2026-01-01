package scorer

import (
	"bean/internal/score"
	"bean/internal/score/rule"
	"bean/internal/trace"
	"context"
	"log/slog"
)

// RulesScorer is a scorer implementation that calculates a score based on a set of rules.
// Each rule evaluates an individual trace, and the resulting scores are accumulated
// within the specified min and max boundaries.
type RulesScorer struct {
	rules []rule.Rule // set of rules to be applied to traces
	min   float32     // minimum allowed value for any score component
	max   float32     // maximum allowed value for any score component
}

// Score computes the final score by applying all rules to each of the provided traces.
// For each trace and rule, rule.Eval is called.
// The resulting deltas are added to the final score, clamped within min and max.
//
// If a rule evaluation fails, the error is logged and the rule is skipped.
// The context is passed for Scorer interface compatibility but is not used.
//
// Returns:
//   - The aggregated final score of type score.Score.
//   - nil as error (rule errors do not halt execution).
func (rs *RulesScorer) Score(ctx context.Context, traces []trace.Trace) (score.Score, error) {
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

// NewRulesScorer creates a new instance of RulesScorer.
// Parameters:
//   - rules: list of rules to apply during scoring
//   - min: minimum value for any score component
//   - max: maximum value for any score component
//
// Typically, min = 0.0 and max = 1.0 to maintain a normalized score range.
//
// Returns a pointer to the initialized scorer.
func NewRulesScorer(rules []rule.Rule, min, max float32) *RulesScorer {
	scorer := RulesScorer{
		rules: rules,
		min:   min,
		max:   max,
	}
	return &scorer
}
