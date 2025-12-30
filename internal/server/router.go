package server

import (
	"bean/internal/dataset"
	"bean/internal/score/scorer"
	"bean/internal/trace"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// ApiV1Router manages routes for API version 1.
// Handles receiving behavioral traces, calculating scores, and serving static files.
// All endpoints follow a REST-like structure.
type ApiV1Router struct {
	// tracesRepo — storage for saving and retrieving behavioral traces by token.
	tracesRepo *trace.TracesRepository
	// todo
	compositeScorer *scorer.CompositeScorer
	// static — path to directory with static files (e.g., collector.js).
	// If empty, static file serving is disabled.
	static string
	// tokenCookie — name of cookie used for session identification when sending traces.
	tokenCookie string
	// datasetRepo - repository for saving behavioral traces
	datasetRepo dataset.DatasetRepository
}

// Mux returns a configured *http.ServeMux with registered handlers.
// Registers the following routes:
// - POST /api/v1/traces — receives new trace
// - GET /api/v1/scores/{token} — retrieves score by token
// - GET /static/... — serves static files (if enabled)
func (ar *ApiV1Router) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/traces", ar.traceHandler)
	mux.HandleFunc("GET /api/v1/scores/{token}", ar.scoreHandler)

	if len(ar.static) != 0 {
		fs := http.FileServer(http.Dir(ar.static))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	}

	return mux
}

// traceHandler handles POST requests with behavioral metrics.
// Expects JSON body with trace data and cookie with name specified in tokenCookie.
// If data is valid, trace is saved to storage.
// On error, returns appropriate HTTP status.
func (ar *ApiV1Router) traceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("Empty trace request body", "error", err, "client", r.RemoteAddr)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	defer r.Body.Close()

	var trace trace.Trace
	err = json.Unmarshal(body, &trace)
	if err != nil {
		slog.Warn("Unable to marshal trace request body", "error", err, "client", r.RemoteAddr)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var token string
	cookies := r.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == ar.tokenCookie {
			token = cookie.Value
			break
		}
	}

	if len(token) == 0 {
		slog.Warn("Empty trace token", "client", r.RemoteAddr)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	slog.Debug("Trace request", "client", r.RemoteAddr, "token", token, "trace", trace)

	ar.tracesRepo.Append(token, trace)
	if ar.datasetRepo != nil {
		ar.datasetRepo.Append(token, trace)
	}
	w.WriteHeader(http.StatusOK)
}

// scoreHandler handles requests to retrieve score by token.
// Token is extracted from URL path: /api/v1/scores/{token}.
// If score is found — returns it in JSON format.
// If not — returns status 404.
func (ar *ApiV1Router) scoreHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if len(token) == 0 {
		slog.Warn("Empty trace token", "client", r.RemoteAddr)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	score, err := ar.compositeScorer.Score(token)
	if err != nil {
		slog.Warn("Score not found", "id", token, "error", err, "client", r.RemoteAddr)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	slog.Debug("Score request", "client", r.RemoteAddr, "token", token, "score", score)

	body, err := json.Marshal(score)
	if err != nil {
		slog.Warn("Unable to marshal score", "error", err, "client", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(body)
}

// NewApiV1Router creates a new API v1 router.
// Parameters:
// - static: path to static files (can be empty)
// - tokenCookie: cookie name for session identification
// - tracesRepo: trace storage
// - scoreCalculator: score calculator
//
// Returns pointer to configured ApiV1Router.
func NewApiV1Router(
	static string,
	tokenCookie string,
	tracesRepo *trace.TracesRepository,
	compositeScorer *scorer.CompositeScorer,
	datasetRepo dataset.DatasetRepository,
) *ApiV1Router {
	return &ApiV1Router{
		tracesRepo:      tracesRepo,
		compositeScorer: compositeScorer,
		static:          static,
		tokenCookie:     tokenCookie,
		datasetRepo:     datasetRepo,
	}
}
