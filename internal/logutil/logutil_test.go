package logutil

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"  info  ", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
	}
	for _, tt := range tests {
		got := ParseLogLevel(tt.in)
		if got != tt.want {
			t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestSplitHandler_InfoToStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitHandler(&stdout, &stderr, slog.LevelInfo, "json")
	logger := slog.New(h)

	logger.Info("hello", "k", "v")
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty; got %q", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("stdout should contain info record")
	}
	if !strings.Contains(stdout.String(), "hello") {
		t.Errorf("stdout missing message: %s", stdout.String())
	}
}

func TestSplitHandler_ErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitHandler(&stdout, &stderr, slog.LevelInfo, "json")
	logger := slog.New(h)

	logger.Error("fail", "err", "something")
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty for error; got %q", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Error("stderr should contain error record")
	}
	if !strings.Contains(stderr.String(), "fail") {
		t.Errorf("stderr missing message: %s", stderr.String())
	}
}

func TestSplitHandler_LevelFilter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitHandler(&stdout, &stderr, slog.LevelWarn, "json")
	logger := slog.New(h)

	logger.Info("hidden")
	if stdout.Len() != 0 {
		t.Errorf("info should be filtered when level is warn; got %q", stdout.String())
	}

	logger.Warn("visible")
	if stderr.Len() != 0 {
		t.Errorf("warn must not go to stderr; got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "visible") {
		t.Errorf("warn should appear on stdout: %s", stdout.String())
	}
}

func TestSplitHandler_WithAttrs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	h := NewSplitHandler(&stdout, &stderr, slog.LevelDebug, "text").WithAttrs([]slog.Attr{slog.String("attr", "value")})
	logger := slog.New(h)

	ctx := context.Background()
	logger.InfoContext(ctx, "msg")
	if !strings.Contains(stdout.String(), "attr=value") {
		t.Errorf("WithAttrs: stdout should contain attr: %s", stdout.String())
	}
}
