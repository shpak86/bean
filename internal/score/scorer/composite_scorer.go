package scorer

import (
	"bean/internal/score"
	"bean/internal/trace"
	"context"
	"errors"
)

// CompositeScorer — агрегирует результаты нескольких scorer'ов, вычисляя итоговую оценку.
// Для заданного идентификатора сессии извлекает трейсы из хранилища и передаёт их
// каждому вложенному scorer'у. Оценки суммируются по ключам с ограничением в диапазоне [0.0, 1.0].
//
// CompositeScorer потокобезопасен, если все вложенные scorer'ы и tracesRepo потокобезопасны.
type CompositeScorer struct {
	scorers    []score.Scorer          // список scorer'ов, чьи оценки будут объединены
	tracesRepo *trace.TracesRepository // хранилище трейсов для получения данных по id
	ctx        context.Context         // контекст, передаваемый scorer'ам при вычислении
}

// Score вычисляет итоговую оценку для указанного идентификатора сессии.
// Порядок действий:
//  1. Получает список трейсов из tracesRepo по id.
//  2. Если трейсы не найдены — возвращает ошибку.
//  3. Вызывает Score у каждого scorer'а, передавая контекст и трейсы.
//  4. Суммирует все оценки по ключам.
//  5. Ограничивает итоговые значения диапазоном [0.0, 1.0].
//
// Возвращает:
//   - Итоговую оценку типа score.Score.
//   - Ошибку, если трейсы не найдены или один из scorer'ов вернул ошибку.
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

// NewCompositeScorer создаёт новый экземпляр CompositeScorer.
// Принимает:
//   - scorers: список scorer'ов, которые будут участвовать в вычислении.
//   - tracesRepo: хранилище трейсов, из которого будут загружаться данные.
//
// Контекст по умолчанию устанавливается как context.Background().
// Для установки кастомного контекста нужно присвоить поле ctx вручную после создания.
func NewCompositeScorer(scorers []score.Scorer, tracesRepo *trace.TracesRepository) *CompositeScorer {
	return &CompositeScorer{scorers: scorers, tracesRepo: tracesRepo}
}
