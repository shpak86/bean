package scorer

import (
	"bean/internal/score"
	"bean/internal/trace"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClientInputScorer is an implementation of a scorer that sends behavioral traces
// to an external ML service for analysis and returns a score.
// Uses HTTP requests with context and timeout.
type ClientInputScorer struct {
	url    string       // URL of the external service for sending traces
	client *http.Client // HTTP client configured with timeout and context cancellation support
	model  string       // model name for prediction
}

// Score sends the provided traces to an external ML service and returns the received score.
// Uses context for request cancellation and timeout.
// Request format: JSON with a "batch" field containing an array of traces.
// The server is expected to return a JSON object with numeric values interpreted as scores.
//
// In case of network error, invalid status (not 200), or incorrect JSON - returns an error.
func (cis *ClientInputScorer) Score(ctx context.Context, traces []trace.Trace) (score.Score, error) {
	requestData := map[string]any{
		"batch": traces,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", cis.url+"/batch", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := cis.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML response error code=%d status=%s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make(score.Score)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// NewClientInputScorer creates a new instance of ClientInputScorer.
// Parameters:
// - url: address of the external ML service (e.g., "http://ml-service:8080/score")
// - timeout: timeout for the HTTP request
//
// Returns a pointer to the initialized scorer.
// Internally uses *http.Client with the specified timeout to manage request duration.
func NewClientInputScorer(url string, timeout time.Duration, model string) *ClientInputScorer {
	client := http.Client{
		Timeout: timeout,
	}

	scorer := &ClientInputScorer{
		url:    url,
		client: &client,
		model:  model,
	}

	return scorer
}
