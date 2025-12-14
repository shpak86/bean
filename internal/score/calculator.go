package score

import (
	"bean/internal/trace"
	"log/slog"
	"gopkg.in/yaml.v3"
)

// ScoreNotFoundError — error returned when no score is found for the specified identifier.
// Occurs in the Score method if the trace repository contains no data for the requested id.
type ScoreNotFoundError struct {
	message string
}

// Error returns the text description of the error.
func (sr *ScoreNotFoundError) Error() string {
	return sr.message
}

// NewScoreNotFoundError creates a new ScoreNotFoundError for the specified identifier.
// Used to indicate absence of behavior data for the given user or session.
func NewScoreNotFoundError(id string) *ScoreNotFoundError {
	return &ScoreNotFoundError{message: "score not found: " + id}
}

// RulesScoreCalculator — component for calculating the final score based on behavioral traces
// and a set of rules defined in a YAML script.
// Each trace is evaluated against all rules, and the results are summed with a limit in the range [0.0, 1.0].
type RulesScoreCalculator struct {
	// tracesRepository — storage of behavioral traces from which data is loaded by identifier.
	tracesRepository *trace.TracesRepository
	// rules — list of rules applied when calculating the score.
	// Rules are processed in declaration order; each can contribute to the final score.
	rules []Rule
}

// Score calculates the final score for the specified identifier (e.g., session or user).
// If traces for id are not found, returns a ScoreNotFoundError.
// Otherwise, the score is calculated by sequentially applying all rules
// to each trace from the history. Results are summed by key, with values limited
// to the range from 0.0 to 1.0 (clamping, not clipping due to saturation).
//
// Rule errors are logged via slog.Error but do not interrupt the calculation.
func (sc *RulesScoreCalculator) Score(id string) (Score, error) {
	traces, found := sc.tracesRepository.Get(id)
	if !found {
		return nil, NewScoreNotFoundError(id)
	}

	score := make(Score)

	for _, trace := range traces {
		for _, rule := range sc.rules {
			delta, err := rule.Eval(trace)
			if err != nil {
				slog.Error("rule eval", "error", err, "rule", rule, "trace", trace)
				continue
			}

			for key, d := range delta {
				newScore := score[key] + d

				switch {
				case newScore < 0.0:
					score[key] = 0.0
				case newScore > 1.0:
					score[key] = 1.0
				default:
					score[key] = newScore
				}
			}
		}
	}

	return score, nil
}

// NewRulesScoreCalculator creates a new score calculator based on a YAML script with rules
// and a reference to the trace storage.
//
// The script should contain a list of rules in the format:
//
// - when: "MouseMoves > 10"
//   then:
//     behavior: 0.5
//
// During creation:
// - Rules are parsed from YAML.
// - For each rule, a CEL program is created and initialized.
//
// In case of syntax errors in YAML or CEL expressions, the corresponding error is returned.
// On successful initialization, returns a pointer to a ready-to-use calculator.
func NewRulesScoreCalculator(script []byte, tracesRepository *trace.TracesRepository) (*RulesScoreCalculator, error) {
	calculator := RulesScoreCalculator{
		tracesRepository: tracesRepository,
		rules:            make([]Rule, 0),
	}

	err := yaml.Unmarshal(script, &calculator.rules)
	if err != nil {
		return nil, err
	}

	for i := range calculator.rules {
		env, err := trace.NewMovementTraceEnv()
		if err != nil {
			return nil, err
		}

		err = calculator.rules[i].Init(env)
		if err != nil {
			return nil, err
		}
	}

	return &calculator, nil
}
