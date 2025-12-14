package score

type Score map[string]float32

type ScoreCalculator interface {
	Score(string) (Score, error)
}
