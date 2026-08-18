package logger

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	opts := Options{
		Filename:   "/tmp/custom.log",
		MaxSize:    50,
		MaxAge:     14,
		MaxBackups: 5,
		Compress:   true,
	}
	lj := New(opts)
	defer func() { _ = lj.Close() }()

	if lj.Filename != "/tmp/custom.log" {
		t.Errorf("expected Filename %q, got %q", "/tmp/custom.log", lj.Filename)
	}
	if lj.MaxSize != 50 {
		t.Errorf("expected MaxSize 50, got %d", lj.MaxSize)
	}
	if lj.MaxAge != 14 {
		t.Errorf("expected MaxAge 14, got %d", lj.MaxAge)
	}
	if lj.MaxBackups != 5 {
		t.Errorf("expected MaxBackups 5, got %d", lj.MaxBackups)
	}
	if !lj.LocalTime {
		t.Errorf("expected LocalTime true, got false")
	}
	if !lj.Compress {
		t.Errorf("expected Compress true, got false")
	}
}

func TestNew_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")
	lj := New(Options{
		Filename:   logPath,
		MaxSize:    10,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	})
	defer func() { _ = lj.Close() }()

	msg := "test log message\n"
	n, err := lj.Write([]byte(msg))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != len(msg) {
		t.Fatalf("expected %d bytes written, got %d", len(msg), n)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(content), "test log message") {
		t.Fatalf("expected log file content to contain %q, got %q", msg, string(content))
	}
}

func TestNewLogger_Stdout(t *testing.T) {
	l, closer := NewLogger(Options{
		Level:  "debug",
		Format: "text",
	})
	if l == nil {
		t.Fatal("expected logger to not be nil")
	}
	if closer != nil {
		t.Errorf("expected closer to be nil when Filename is empty, got non-nil")
	}
}

func TestNewLogger_TextFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "text.log")
	l, closer := NewLogger(Options{
		Level:      "info",
		Format:     "text",
		Filename:   logPath,
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 7,
		Compress:   true,
	})
	if l == nil {
		t.Fatal("expected logger to not be nil")
	}
	if closer == nil {
		t.Fatal("expected closer to not be nil when Filename is set")
	}

	l.Info("slog text file test", "key", "value")

	if err := closer(); err != nil {
		t.Fatalf("closer error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(content), "slog text file test") {
		t.Fatalf("expected log file to contain log message, got %q", string(content))
	}
	if !strings.Contains(string(content), "key=value") {
		t.Fatalf("expected log file to contain structured attribute key=value, got %q", string(content))
	}
}

func TestNewLogger_JSONFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "json.log")
	l, closer := NewLogger(Options{
		Level:      "warn",
		Format:     "json",
		Filename:   logPath,
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 7,
		Compress:   true,
	})
	if l == nil {
		t.Fatal("expected logger to not be nil")
	}
	if closer == nil {
		t.Fatal("expected closer to not be nil when Filename is set")
	}

	l.Warn("slog json test", "module", "auth")

	if err := closer(); err != nil {
		t.Fatalf("closer error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("expected valid JSON in log output, got unmarshal error: %v; raw: %s", err, string(content))
	}
	if parsed["msg"] != "slog json test" {
		t.Errorf("expected msg 'slog json test', got %v", parsed["msg"])
	}
	if parsed["module"] != "auth" {
		t.Errorf("expected module 'auth', got %v", parsed["module"])
	}
	if parsed["level"] != "WARN" {
		t.Errorf("expected level 'WARN', got %v", parsed["level"])
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.want {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
