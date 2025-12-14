package score

import (
	"bean/internal/trace"
	"log/slog"

	"gopkg.in/yaml.v3"
)

// ScoreNotFoundError — ошибка, возвращаемая, когда для указанного идентификатора не найдено оценки.
// Возникает в методе Score, если в репозитории трейсов отсутствуют данные для запрошенного id.
type ScoreNotFoundError struct {
	message string
}

// Error возвращает текстовое описание ошибки.
func (sr *ScoreNotFoundError) Error() string {
	return sr.message
}

// NewScoreNotFoundError создаёт новую ошибку типа ScoreNotFoundError для указанного идентификатора.
// Используется для обозначения отсутствия данных о поведении для данного пользователя или сессии.
func NewScoreNotFoundError(id string) *ScoreNotFoundError {
	return &ScoreNotFoundError{message: "score not found: " + id}
}

// RulesScoreCalculator — компонент для вычисления итоговой оценки на основе поведенческих трейсов
// и набора правил, определённых в YAML-скрипте.
// Каждый трейс оценивается по всем правилам, а результаты суммируются с ограничением в диапазоне [0.0, 1.0].
type RulesScoreCalculator struct {
	// tracesRepository — хранилище поведенческих трейсов, откуда загружаются данные по идентификатору.
	tracesRepository *trace.TracesRepository

	// rules — список правил, применяемых при вычислении оценки.
	// Правила обрабатываются в порядке объявления; каждое может внести вклад в итоговую оценку.
	rules []Rule
}

// Score вычисляет итоговую оценку для указанного идентификатора (например, сессии или пользователя).
// Если трейсы для id не найдены, возвращается ошибка ScoreNotFoundError.
// В противном случае оценка вычисляется путём последовательного применения всех правил
// к каждому трейсу из истории. Результаты суммируются по ключам, при этом значения ограничиваются
// диапазоном от 0.0 до 1.0 (усечение, а не обрезание за счёт насыщения).
//
// Логирование ошибок правил осуществляется через slog.Error, но не прерывает вычисление.
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

// NewRulesScoreCalculator создаёт новый калькулятор оценок на основе YAML-скрипта с правилами
// и ссылки на хранилище трейсов.
//
// Скрипт должен содержать список правил в формате:
//
//   - when: "MouseMoves > 10"
//     then:
//     behavior: 0.5
//
// При создании:
//   - Правила разбираются из YAML.
//   - Для каждого правила создаётся и инициализируется CEL-программа.
//
// В случае синтаксических ошибок в YAML или CEL-выражениях возвращается соответствующая ошибка.
// При успешной инициализации возвращается указатель на готовый к использованию калькулятор.
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
