// Package logger — writer log dengan rotasi otomatis menggunakan lumberjack
// dan utilitas pembentukan logger standard library slog.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Options mendefinisikan opsi konfigurasi untuk logger.
type Options struct {
	Level      string // debug, info, warn, error
	Format     string // text, json
	Filename   string // file path (kosong = stdout)
	MaxSize    int    // megabytes
	MaxAge     int    // days
	MaxBackups int    // count
	Compress   bool   // gzip compression
}

// New membuat *lumberjack.Logger untuk rolling log file berdasarkan Options.
func New(opts Options) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   opts.Filename,
		MaxSize:    opts.MaxSize,
		MaxAge:     opts.MaxAge,
		MaxBackups: opts.MaxBackups,
		LocalTime:  true,
		Compress:   opts.Compress,
	}
}

// NewLogger membuat *slog.Logger berdasarkan Options. Jika `Filename` tidak kosong,
// log akan ditulis ke rolling file lumberjack. Jika kosong, log ditulis ke stdout.
// Fungsi mengembalikan logger dan fungsi cleanup/closer (jika ada file yang dibuka).
func NewLogger(opts Options) (*slog.Logger, func() error) {
	var logOut io.Writer = os.Stdout
	var closer func() error

	if opts.Filename != "" {
		lj := New(opts)
		logOut = lj
		closer = lj.Close
	}

	handlerOpts := &slog.HandlerOptions{
		Level: ParseLevel(opts.Level),
	}

	var handler slog.Handler
	if strings.EqualFold(strings.TrimSpace(opts.Format), "json") {
		handler = slog.NewJSONHandler(logOut, handlerOpts)
	} else {
		handler = slog.NewTextHandler(logOut, handlerOpts)
	}

	logger := slog.New(handler)
	return logger, closer
}

// ParseLevel memetakan string ("debug", "info", "warn", "error") ke slog.Level.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
