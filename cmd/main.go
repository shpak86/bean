package main

import (
	"bean/internal/configuration"
	"bean/internal/score"
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

// prepareLogger настраивает глобальный логгер с использованием slog.
// Принимает строковый уровень логирования (например, "debug", "info", "warn", "error")
// и устанавливает JSON-форматированный вывод на os.Stdout.
// Если уровень не распознан, используется уровень Info по умолчанию.
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

// При ошибках на этапе загрузки конфигурации, чтения правил или инициализации компонентов
// приложение завершается с кодом 1.
func main() {
	configPath := flag.String("config", "/etc/bean/config.yaml", "configuration file")
	flag.Parse()
	config, err := configuration.LoadConfig(*configPath)
	if err != nil {
		slog.Error("Unable to load configuration", "error", err)
		os.Exit(1)
	}
	prepareLogger(config.Logger.Level)

	appCtx, appCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appCancel()

	tracesRepo := trace.NewTracesRepository(config.Analysis.TracesLength, config.Analysis.TracesTtl)
	go tracesRepo.Serve()

	content, err := os.ReadFile(config.Analysis.Rules)
	if err != nil {
		slog.Error("Unable to load rules", "error", err)
		os.Exit(1)
	}
	scoreCalc, err := score.NewRulesScoreCalculator(content, tracesRepo)
	if err != nil {
		slog.Error("Unable to initialize score calculator", "error", err)
		os.Exit(1)
	}
	srv := server.NewServer(
		config.Server.Address,
		config.Server.Static,
		config.Analysis.Token,
		tracesRepo,
		scoreCalc,
	)
	go srv.ListenAndServe()
	slog.Info("Server listening " + config.Server.Address)
	<-appCtx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownCancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("Server shutdown", "error", err)
	}
	slog.Info("Server stopped")

	tracesRepo.Stop()
}
