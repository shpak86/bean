package server

import (
	"bean/internal/score"
	"bean/internal/trace"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type ApiV1Router struct {
	tracesRepo      *trace.TracesRepository
	scoreCalculator *score.RulesScoreCalculator
	static          string
	tokenCookie     string
}

func (ar *ApiV1Router) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/trace", ar.traceHandler)
	mux.HandleFunc("GET /api/v1/score", ar.scoreHandler)
	if len(ar.static) != 0 {
		fs := http.FileServer(http.Dir(ar.static))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	}
	return mux
}

func (ar *ApiV1Router) traceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("Empty trace request body", "error", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()

	var trace trace.Trace
	err = json.Unmarshal(body, &trace)
	if err != nil {
		slog.Warn("Unable to marshal trace request body", "error", err)
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
		slog.Warn("Empty trace token")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	ar.tracesRepo.Append(token, trace)
	w.WriteHeader(http.StatusOK)
}

func (ar *ApiV1Router) scoreHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(ar.tokenCookie)
	if len(token) == 0 {
		slog.Warn("Empty trace token")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	score, err := ar.scoreCalculator.Score(token)
	if err != nil {
		slog.Warn("Score not found", "id", token, "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(score)
	if err != nil {
		slog.Warn("Unable to marshal score", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(body)
}

func NewApiV1Router(
	static string,
	tokenCookie string,
	tracesRepo *trace.TracesRepository,
	scoreCalculator *score.RulesScoreCalculator,
) *ApiV1Router {
	return &ApiV1Router{
		tracesRepo:      tracesRepo,
		scoreCalculator: scoreCalculator,
		static:          static,
		tokenCookie:     tokenCookie,
	}
}
