package server

import (
	"bean/internal/score"
	"bean/internal/trace"
	"context"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewServer(
	address string,
	static string,
	tokenCookie string,
	tracesRepo *trace.TracesRepository,
	scoreCalculator *score.RulesScoreCalculator,
) *Server {
	router := NewApiV1Router(static, tokenCookie, tracesRepo, scoreCalculator)
	s := Server{&http.Server{
		Addr:           address,
		Handler:        router.Mux(),
		ReadTimeout:    time.Second * 3,
		WriteTimeout:   time.Second * 3,
		MaxHeaderBytes: 1024 * 10,
	}}
	return &s
}
