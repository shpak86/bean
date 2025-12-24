package dataset

import (
	"bean/internal/trace"
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

// customJSONHandler is a custom slog handler that outputs logs in JSON format
// with time in "2006-01-01 15:04:05" format and without the log level field.
// All attributes are written at the top level of the object.
type customJSONHandler struct {
	opts slog.HandlerOptions // handler options (not actively used, but stored)
	out  io.Writer           // target writer for JSON record output
}

// NewCustomJSONHandler creates a new instance of CustomJSONHandler.
// Parameters:
// - out: writer where JSON logs will be written (e.g., file)
// - opts: slog.HandlerOptions (can be nil)
//
// Returns a ready-to-use CustomJSONHandler.
func NewCustomJSONHandler(out io.Writer, opts *slog.HandlerOptions) *customJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &customJSONHandler{
		opts: *opts,
		out:  out,
	}
}

// Handle implements the slog.Handler interface: serializes a record to JSON
// with the required time format and without the log level.
// Each record is written as a separate line (JSONL format).
func (h *customJSONHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]interface{})

	// Set time in the required format
	attrs["time"] = r.Time.Format("2006-01-01 15:04:05")

	// Add all record attributes
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != "" && a.Value.Any() != nil {
			attrs[a.Key] = a.Value.Any()
		}

		return true
	})

	// Serialize to JSON
	data, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	// Write to writer with newline
	_, err = h.out.Write(append(data, '\n'))
	return err
}

// WithAttrs is not supported
func (h *customJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("WithAttrs is not supported by CustomJSONHandler")
}

// WithGroup is not supported
func (h *customJSONHandler) WithGroup(name string) slog.Handler {
	panic("WithGroup is not supported by CustomJSONHandler")
}

// Enabled determines whether the handler should process a record of the given level.
// Always returns true — all levels are allowed.
func (h *customJSONHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// JsonDatasetRepository is a thread-safe repository for collecting behavioral trace datasets.
// Writes each trace to a JSON file with rotation and compression via lumberjack.
// Suitable for long-term data collection.
type JsonDatasetRepository struct {
	lumberjack *lumberjack.Logger // rotating file logger
	logger     *slog.Logger       // structured logger with custom output
}

// NewJsonDatasetRepository creates a new repository for dataset collection.
// Parameters:
// - file: path to the file where data is written
// - maxSize: maximum file size in MB before rotation
// - maxBackups: maximum number of old files to keep
//
// Returns a pointer to an initialized repository.
func NewJsonDatasetRepository(file string, maxSize, maxBackups int) *JsonDatasetRepository {
	repo := JsonDatasetRepository{}
	repo.lumberjack = &lumberjack.Logger{
		Filename:   file,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		Compress:   true,
	}

	handler := NewCustomJSONHandler(repo.lumberjack, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	repo.logger = slog.New(handler)
	return &repo
}

// Append adds a trace to the dataset with binding to the session token.
// Recording occurs as a JSON object with "token" and "trace" fields.
// The method is thread-safe thanks to lumberjack and slog.
func (r *JsonDatasetRepository) Append(token string, t trace.Trace) {
	r.logger.Info("", "token", token, "trace", t)
}

// Close closes the underlying file. Should be called when shutting down
// to ensure write completion and rotation of the last file.
func (r *JsonDatasetRepository) Close() {
	r.lumberjack.Close()
}
