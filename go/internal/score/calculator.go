package score

import (
	"bean/internal/trace"
	"log/slog"

	"gopkg.in/yaml.v3"
)

type ScoreNotFoundError struct {
	message string
}

func (sr *ScoreNotFoundError) Error() string {
	return sr.message
}

func NewScoreNotFoundError(id string) *ScoreNotFoundError {
	return &ScoreNotFoundError{message: "score not found: " + id}
}

type RulesScoreCalculator struct {
	tracesRepository *trace.TracesRepository
	rules            []Rule
}

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
