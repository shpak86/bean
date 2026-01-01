package rule

import (
	"bean/internal/score"
	"bean/internal/trace"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRule_Init_Success(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("MouseMoves", cel.IntType),
	)
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > 10",
	}

	err = rule.Init(env)
	assert.NoError(t, err)
	assert.NotNil(t, rule.program, "program should be compiled and assigned")
}

func TestRule_Init_ParseError(t *testing.T) {
	env, err := cel.NewEnv()
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > ", // invalid syntax
	}

	err = rule.Init(env)
	assert.Error(t, err, "expected parse error for invalid expression")
}

func TestRule_Init_CheckError(t *testing.T) {
	env, err := cel.NewEnv()
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > '10'", // type mismatch: comparing int and string
	}

	err = rule.Init(env)
	assert.Error(t, err, "expected check error for type mismatch")
}

func TestRule_Eval_TrueCondition(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("MouseMoves", cel.IntType),
	)
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > 5",
		Then: score.Score{"behavior": 0.5},
	}

	err = rule.Init(env)
	require.NoError(t, err)

	tt := trace.Trace{"MouseMoves": int32(10)}
	s, err := rule.Eval(tt)

	assert.NoError(t, err)
	assert.Equal(t, score.Score{"behavior": 0.5}, s, "should return Then score when condition is true")
}

func TestRule_Eval_FalseCondition(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("MouseMoves", cel.IntType),
	)
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > 10",
		Then: score.Score{"behavior": 0.5},
	}

	err = rule.Init(env)
	require.NoError(t, err)

	tt := trace.Trace{"MouseMoves": int32(5)}
	score, err := rule.Eval(tt)

	assert.NoError(t, err)
	assert.Empty(t, score, "should return empty")
}

func TestRule_Eval_UndefinedField(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("Clicks", cel.IntType),
	)
	require.NoError(t, err)

	rule := &Rule{
		When: "Clicks > 3",
		Then: score.Score{"behavior": 0.7},
	}

	err = rule.Init(env)
	require.NoError(t, err)

	// Trace doesn't contain 'Clicks' — in CEL this would be error, but we pass map[string]any
	tt := trace.Trace{"MouseMoves": int32(10)}
	score, err := rule.Eval(tt)

	assert.NoError(t, err)
	assert.Empty(t, score, "should return empty")
}

func TestRule_Eval_ComplexCondition(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("MouseMoves", cel.IntType),
		cel.Variable("Clicks", cel.IntType),
		cel.Variable("Scrolls", cel.IntType),
	)
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > 5 && (Clicks > 2 || Scrolls > 1)",
		Then: score.Score{"behavior": 0.9},
	}

	err = rule.Init(env)
	require.NoError(t, err)

	// Condition: MouseMoves > 5 and (Clicks > 2 or Scrolls > 1) → true
	tt := trace.Trace{
		"MouseMoves": int32(10),
		"Clicks":     int32(1),
		"Scrolls":    int32(2),
	}

	s, err := rule.Eval(tt)

	assert.NoError(t, err)
	assert.Equal(t, score.Score{"behavior": 0.9}, s, "should evaluate complex condition correctly")
}

func TestRule_Eval_NilTrace(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("MouseMoves", cel.IntType),
	)
	require.NoError(t, err)

	rule := &Rule{
		When: "MouseMoves > 5",
		Then: score.Score{"behavior": 0.5},
	}

	err = rule.Init(env)
	require.NoError(t, err)

	var tt trace.Trace // nil map
	score, err := rule.Eval(tt)

	assert.NoError(t, err)
	assert.Empty(t, score, "should return empty")
}
