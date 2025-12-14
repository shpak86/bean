package score

import (
	"testing"

	"bean/internal/trace"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRulesScoreCalculator_UnmarshalError(t *testing.T) {
	invalidYAML := []byte("when: invalid yaml [[[[[")
	repo := trace.NewTracesRepository(2, 0)

	_, err := NewRulesScoreCalculator(invalidYAML, repo)
	assert.Error(t, err, "expected error for invalid YAML")
}

func TestNewRulesScoreCalculator_InitRuleError(t *testing.T) {
	// CEL выражение с неизвестной переменной, если окружение её не объявляет
	// Предположим, что trace.NewMovementTraceEnv() предоставляет только основные поля
	script := `
- when: "unknownField == 1"
  then:
    behavior: 0.5
`
	repo := trace.NewTracesRepository(2, 0)

	_, err := NewRulesScoreCalculator([]byte(script), repo)
	// Ожидаем ошибку компиляции правила
	assert.Error(t, err, "expected error when rule uses undefined variable")
}

func TestRulesScoreCalculator_Score_TraceNotFound(t *testing.T) {
	repo := trace.NewTracesRepository(2, 0)
	script := `[]`
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	score, err := calculator.Score("unknown")
	assert.Nil(t, score)
	assert.Error(t, err)
	var notFound *ScoreNotFoundError
	assert.ErrorAs(t, err, &notFound, "error should be ScoreNotFoundError")
	assert.Contains(t, err.Error(), "score not found: unknown")
}

func TestRulesScoreCalculator_Score_NoRules_NoScore(t *testing.T) {
	repo := trace.NewTracesRepository(2, 0)
	script := `[]`
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	// Добавим трейс
	repo.Append("user1", trace.Trace{"mouseMoves": int32(10)})
	score, err := calculator.Score("user1")
	assert.NoError(t, err)
	assert.Equal(t, Score{}, score, "should return empty score when no rules")
}

func TestRulesScoreCalculator_Score_RuleApplies(t *testing.T) {
	const script = `
- when: mouseMoves > 5
  then:
    behavior: 0.5
`
	repo := trace.NewTracesRepository(2, 0)
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	// Добавим трейс, который удовлетворяет условию
	repo.Append("user1", trace.Trace{"mouseMoves": int32(10)})

	score, err := calculator.Score("user1")
	assert.NoError(t, err)
	assert.Equal(t, Score{"behavior": 0.5}, score, "should apply rule and return correct score")
}

func TestRulesScoreCalculator_Score_RuleDoesNotApply(t *testing.T) {
	const script = `
- when: mouseMoves > 10
  then:
    behavior: 0.5
`
	repo := trace.NewTracesRepository(2, 0)
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	// Добавим трейс, который НЕ удовлетворяет условию
	repo.Append("user1", trace.Trace{"mouseMoves": int32(5)})

	score, err := calculator.Score("user1")
	assert.NoError(t, err)
	assert.Equal(t, Score{}, score, "should return empty score when no rule matches")
}

func TestRulesScoreCalculator_Score_MultipleTracesAndRules(t *testing.T) {
	const script = `
- when: mouseMoves > 5
  then:
    behavior: 0.8
- when: clicks > 2 && clicks <= 5
  then:
    behavior: 0.3
`
	repo := trace.NewTracesRepository(10, 0)
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	// Добавим несколько трейсов
	repo.Append("user1", trace.Trace{"mouseMoves": int32(10), "clicks": int32(1)})
	repo.Append("user1", trace.Trace{"mouseMoves": int32(2), "clicks": int32(3)})
	repo.Append("user1", trace.Trace{"mouseMoves": int32(4), "clicks": int32(1)})

	score, err := calculator.Score("user1")
	assert.NoError(t, err)
	// Первый и третий трейс: +0.8 (mouseMoves > 5)
	// Второй трейс: +0.3 (clicks = 3)
	// Третий трейс: +0.3 (clicks > 2 && clicks <= 5)
	// Сумма: 1.0, не выходит за пределы [0,1]
	assert.Equal(t, Score{"behavior": 1.0}, score, "should accumulate score from multiple matching rules and traces")
}

func TestRulesScoreCalculator_Score_ScoreClamping(t *testing.T) {
	const script = `
- when: mouseMoves > 1
  then:
    behavior: 0.8
- when: true
  then:
    behavior: 0.5
`
	repo := trace.NewTracesRepository(2, 0)
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	// Два трейса: каждый сработает по обоим правилам → 2 * (0.8 + 0.5) = 2.6 → должно быть обрезано до 1.0
	repo.Append("user1", trace.Trace{"mouseMoves": int32(10)})
	repo.Append("user1", trace.Trace{"mouseMoves": int32(10)})

	score, err := calculator.Score("user1")
	assert.NoError(t, err)
	assert.Equal(t, Score{"behavior": 1.0}, score, "score should be clamped to 1.0")
}

func TestRulesScoreCalculator_Score_MultipleDimensions(t *testing.T) {
	const script = `
- when: mouseMoves > 5
  then:
    automation: 0.5
- when: clicks > 1
  then:
    automation: 0.3
    behavior: 0.2
`
	repo := trace.NewTracesRepository(10, 0)
	calculator, err := NewRulesScoreCalculator([]byte(script), repo)
	require.NoError(t, err)

	repo.Append("user1", trace.Trace{"mouseMoves": int32(10), "clicks": int32(2)})

	score, err := calculator.Score("user1")
	assert.NoError(t, err)
	assert.Equal(t, float32(0.5+0.3), score["automation"], "automation score should sum correctly")
	assert.Equal(t, float32(0.2), score["behavior"], "behavior score should be set")
}

func TestNewRulesScoreCalculator_EmptyScript(t *testing.T) {
	repo := trace.NewTracesRepository(2, 0)
	script := []byte("")
	calculator, err := NewRulesScoreCalculator(script, repo)
	require.NoError(t, err)
	assert.Empty(t, calculator.rules, "should handle empty script as empty rules")
}

func TestNewRulesScoreCalculator_ValidYAMLInvalidStructure(t *testing.T) {
	// Валидный YAML, но не соответствует []Rule
	script := []byte("not: a list")
	repo := trace.NewTracesRepository(2, 0)

	_, err := NewRulesScoreCalculator(script, repo)
	assert.Error(t, err, "should fail to unmarshal into []Rule")
}
