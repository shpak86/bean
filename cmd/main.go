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

// prepareScorers creates a list of scorers
// Accepts list of scorers configurations.
// Returns list of scorers.
func prepareScorers(sc []configuration.ScorerConfig) []score.TracesScorer {
	scorers := []score.TracesScorer{}
	for i := range sc {
		switch sc[i].Type {
		case configuration.ScorerTypeML:
			mlScorer := scorer.NewClientInputScorer(sc[i].Url, time.Second, sc[i].Model)
			scorers = append(scorers, mlScorer)
		case configuration.ScorerTypeRules:
			rules, err := rule.LoadFromFile(sc[i].Rules, trace.NewMovementTraceEnv)
			if err != nil {
				slog.Error("Unable to load rules", "file", sc[i].Rules, "error", err)
				os.Exit(1)
			}
			rulesScorer := scorer.NewRulesScorer(rules, -1.0, 1.0)
			scorers = append(scorers, rulesScorer)
		default:
			slog.Error("Unknown scorer", "scorer", sc[i].Type)
			os.Exit(1)
		}
	}
	return scorers
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

	scorers := prepareScorers(config.Analysis.Scorers)
	compositeScorer := scorer.NewCompositeScorer(scorers, tracesRepo)

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
