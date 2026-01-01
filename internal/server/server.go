package server

import (
	"bean/internal/dataset"
	"bean/internal/score/scorer"
	"bean/internal/trace"
	"context"
	"net/http"
	"time"
)

// Server encapsulates the HTTP server of the application, providing controlled startup and shutdown.
// Uses a customizable router and ensures timeouts for security and stability.
type Server struct {
	// server — embedded HTTP server from net/http package, fully configured and ready to use.
	server *http.Server
}

// ListenAndServe starts the HTTP server and begins listening on the specified address.
// Blocks execution until the server is stopped or an error occurs.
// If server is stopped via Shutdown, method returns http.ErrServerClosed.
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server with the provided context.
// Stops listening, terminates accepting new connections, and allows active connections
// to complete within the timeout specified in the context.
// Should be called during graceful shutdown of the application.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// NewServer creates and configures a new server instance.
//
// Parameters:
// - address: address and port to listen on (e.g., ":8080").
// - static: path to directory with static files to be served.
// - tokenCookie: name of cookie used for request authentication.
// - tracesRepo: repository for storing and retrieving behavioral traces.
// - scoreCalculator: calculator used for computing scores based on traces.
// - datasetRepo: repository for storing bahavioral traces
//
// Configures API v1 routes, including static file handling and behavioral metrics processing.
// Sets secure timeouts for reading and writing, and limits header size.
//
// Returns pointer to a ready-to-run server.
func NewServer(
	address string,
	static string,
	tokenCookie string,
	tracesRepo *trace.TracesRepository,
	compositeScorer *scorer.CompositeScorer,
	datasetRepo dataset.DatasetRepository,
) *Server {
	router := NewApiV1Router(static, tokenCookie, tracesRepo, compositeScorer, datasetRepo)
	s := Server{&http.Server{
		Addr:           address,
		Handler:        router.Mux(),
		ReadTimeout:    time.Second * 3,
		WriteTimeout:   time.Second * 3,
		MaxHeaderBytes: 1024 * 10,
	}}

	return &s
}
