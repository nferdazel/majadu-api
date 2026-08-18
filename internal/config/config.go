// Package config — konfigurasi runtime dari environment; `.env` di-load
// untuk dev lokal, prod memakai systemd/podman env.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config — runtime configuration. Semua nilai dari env; `.env` di-load
// untuk dev lokal (godotenv). Prod memakai systemd/podman env (file 600).
type Config struct {
	// Env — "dev" | "prod". Dev = lenient (default fallback), prod = strict.
	Env string `env:"MAJADU_ENV" envDefault:"prod"`

	// Port HTTP server. Default 8080.
	Port string `env:"PORT" envDefault:"8080"`

	// DatabaseURL — postgres://... koneksi ke Postgres.
	// WAJIB diisi (fail-fast).
	DatabaseURL string `env:"DATABASE_URL,required"`

	// DatabaseSchema — schema Postgres yang dipakai (bm / bm_dev).
	// WAJIB: di-set via env, BUKAN hardcode — supaya merge dev→main tidak
	// membawa schema dev ke prod. Dev: bm_dev. Prod: bm.
	DatabaseSchema string `env:"MAJADU_DB_SCHEMA,required"`

	// AllowedOrigins — daftar origin CORS yang diizinkan (frontend Vercel).
	// Bisa comma-separated; "*" = semua (hanya untuk dev).
	AllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:","`

	// BaseURL — URL publik API (untuk header Location), mis.
	// https://api.qouver.com/majadu/v1. Kosong = relative path.
	BaseURL string `env:"PUBLIC_BASE_URL"`

	// RateLimitPerMin — batas request per menit per IP. 0 = disabled.
	RateLimitPerMin int `env:"MAJADU_RATE_LIMIT_PER_MIN" envDefault:"120"`

	// LogLevel — "debug" | "info" | "warn" | "error". Default info.
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// LogFormat — format log: "text" | "json". Default text.
	LogFormat string `env:"LOG_FORMAT" envDefault:"text"`

	// LogFile — file path target log (mis. /logs/app.log). Kosong = stdout saja.
	LogFile string `env:"LOG_FILE"`

	// LogMaxSize — ukuran maksimal per file log dalam MB sebelum dirotasi. Default 100.
	LogMaxSize int `env:"LOG_MAX_SIZE" envDefault:"100"`

	// LogMaxAge — masa retensi log file lama dalam hari. Default 7.
	LogMaxAge int `env:"LOG_MAX_AGE" envDefault:"7"`

	// LogMaxBackups — jumlah maksimal file cadangan log yang disimpan. Default 7.
	LogMaxBackups int `env:"LOG_MAX_BACKUPS" envDefault:"7"`

	// LogCompress — kompresi file cadangan log dengan gzip. Default true.
	LogCompress bool `env:"LOG_COMPRESS" envDefault:"true"`
}

// Load membaca env, me-load .env jika ada, lalu validasi.
// Prod gagal cepat (exit) kalau config wajib tidak ada.
func Load() (Config, error) {
	_ = godotenv.Load()
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
