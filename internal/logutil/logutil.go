package logutil

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// SplitHandler routes log records by level: Debug, Info, and Warn go to stdoutWriter;
// Error and above go to stderrWriter.
type SplitHandler struct {
	stdout slog.Handler
	stderr slog.Handler
}

// NewSplitHandler returns a handler that writes records with level < Error to stdoutWriter
// and records with level >= Error to stderrWriter. format is "json" or "text".
// stdoutWriter and stderrWriter must be non-nil.
func NewSplitHandler(stdoutWriter, stderrWriter io.Writer, level slog.Level, format string) *SplitHandler {
	opts := &slog.HandlerOptions{Level: level}
	stderrOpts := &slog.HandlerOptions{Level: slog.LevelError}
	var stdout, stderr slog.Handler
	if strings.ToLower(strings.TrimSpace(format)) == "json" {
		stdout = slog.NewJSONHandler(stdoutWriter, opts)
		stderr = slog.NewJSONHandler(stderrWriter, stderrOpts)
	} else {
		stdout = slog.NewTextHandler(stdoutWriter, opts)
		stderr = slog.NewTextHandler(stderrWriter, stderrOpts)
	}
	return &SplitHandler{stdout: stdout, stderr: stderr}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SplitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level)
}

// Handle routes the record to stdout (Debug/Info/Warn) or stderr (Error+).
func (h *SplitHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		return h.stderr.Handle(ctx, r)
	}
	return h.stdout.Handle(ctx, r)
}

// WithAttrs returns a new handler with the given attributes added to both inner handlers.
func (h *SplitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SplitHandler{
		stdout: h.stdout.WithAttrs(attrs),
		stderr: h.stderr.WithAttrs(attrs),
	}
}

// WithGroup returns a new handler with the given group name for both inner handlers.
func (h *SplitHandler) WithGroup(name string) slog.Handler {
	return &SplitHandler{
		stdout: h.stdout.WithGroup(name),
		stderr: h.stderr.WithGroup(name),
	}
}

// ParseLogLevel returns slog.Level from a string (debug, info, warn, error). Defaults to info.
func ParseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
