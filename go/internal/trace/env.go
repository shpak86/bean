package trace

import "github.com/google/cel-go/cel"

func NewMovementTraceEnv() (*cel.Env, error) {
	env, err := cel.NewEnv(
		// --- Timestamp ---
		cel.Variable("timestamp", cel.StringType),

		// --- Behavior Metrics ---
		cel.Variable("mouseMoves", cel.IntType),
		cel.Variable("clicks", cel.IntType),
		cel.Variable("clickTimingMin", cel.IntType),
		cel.Variable("clickTimingMax", cel.IntType),
		cel.Variable("clickTimingAvg", cel.IntType),
		cel.Variable("clickTimingCount", cel.IntType),
		cel.Variable("scrolls", cel.IntType),
		cel.Variable("scrollTimingMin", cel.IntType),
		cel.Variable("scrollTimingMax", cel.IntType),
		cel.Variable("scrollTimingAvg", cel.IntType),
		cel.Variable("scrollTimingCount", cel.IntType),
		cel.Variable("textInputEvents", cel.IntType),
		cel.Variable("textInputTimingMin", cel.IntType),
		cel.Variable("textInputTimingMax", cel.IntType),
		cel.Variable("textInputTimingAvg", cel.IntType),
		cel.Variable("textInputTimingCount", cel.IntType),
		cel.Variable("sessionDuration", cel.IntType),

		// --- Browser and Device Info ---
		cel.Variable("userAgent", cel.StringType),
		cel.Variable("language", cel.StringType),
		cel.Variable("platform", cel.StringType),
		cel.Variable("screenWidth", cel.IntType),
		cel.Variable("screenHeight", cel.IntType),
		cel.Variable("timezone", cel.StringType),
		cel.Variable("cookiesEnabled", cel.BoolType),
		cel.Variable("onLine", cel.BoolType),
		cel.Variable("deviceMemory", cel.IntType),
		cel.Variable("maxTouchPoints", cel.IntType),
		cel.Variable("browserName", cel.StringType),
		cel.Variable("browserVersion", cel.StringType),
		cel.Variable("osName", cel.StringType),
		cel.Variable("osVersion", cel.StringType),
	)
	if err != nil {
		return nil, err
	}
	return env, nil
}
