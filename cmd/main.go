package main

import (
	"bean/internal/configuration"
	"bean/internal/dataset"
	"bean/internal/score"
	"bean/internal/score/rule"
	"bean/internal/score/scorer"
	"bean/internal/server"
	"bean/internal/trace"
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// prepareLogger configures the global logger using slog.
// Accepts a log level string (e.g., "debug", "info", "warn", "error")
// and sets up JSON-formatted output to os.Stdout.
// If the level is not recognized, the Info level is used by default.
func prepareLogger(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// On errors during config loading, rules reading, or component initialization,
// the application exits with code 1.
func main() {
	configPath := flag.String("config", "/etc/bean/config.yaml", "configuration file")
	flag.Parse()

	config, err := configuration.LoadConfig(*configPath)
	if err != nil {
		slog.Error("Unable to load configuration", "error", err)
		os.Exit(1)
	}

	prepareLogger(config.Logger.Level)
	var datasetRepo dataset.DatasetRepository
	if config.Dataset.File != "" {
		datasetRepo = dataset.NewJsonDatasetRepository(config.Dataset.File, config.Dataset.Size, config.Dataset.Amount)
	}

	appCtx, appCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appCancel()

	tracesRepo := trace.NewTracesRepository(config.Analysis.TracesLength, config.Analysis.TracesTtl)
	go tracesRepo.Serve()

	rules, err := rule.LoadFromFile(config.Analysis.Rules, trace.NewMovementTraceEnv)
	if err != nil {
		slog.Error("Unable to load rules", "file", config.Analysis.Rules, "error", err)
		os.Exit(1)
	}
	rulesScorer := scorer.NewRulesScorer(rules, -1.0, 1.0)
	mlScorer := scorer.NewClientInputScorer("http://127.0.0.1:8000/batch", time.Minute)

	compositeScorer := scorer.NewCompositeScorer([]score.TracesScorer{mlScorer, rulesScorer}, tracesRepo)
	if err != nil {
		slog.Error("Unable to initialize score calculator", "error", err)
		os.Exit(1)
	}

	srv := server.NewServer(
		config.Server.Address,
		config.Server.Static,
		config.Analysis.Token,
		tracesRepo,
		compositeScorer,
		datasetRepo,
	)

	go srv.ListenAndServe()
	slog.Info("Server is listening " + config.Server.Address)

	<-appCtx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownCancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("Server shutdown", "error", err)
	}

	slog.Info("Server stopped")
	tracesRepo.Stop()
	datasetRepo.Close()
}
