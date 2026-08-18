package config_test

import (
	"os"
	"reflect"
	"testing"

	"majadu-api/internal/config"
)

// clearEnv unsets all config-related environment variables for test isolation.
func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"MAJADU_ENV",
		"PORT",
		"DATABASE_URL",
		"MAJADU_DB_SCHEMA",
		"CORS_ALLOWED_ORIGINS",
		"PUBLIC_BASE_URL",
		"MAJADU_RATE_LIMIT_PER_MIN",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"LOG_FILE",
		"LOG_MAX_SIZE",
		"LOG_MAX_AGE",
		"LOG_MAX_BACKUPS",
		"LOG_COMPRESS",
	}
	for _, k := range keys {
		if old, ok := os.LookupEnv(k); ok {
			os.Unsetenv(k)
			t.Cleanup(func() {
				os.Setenv(k, old)
			})
		}
	}
}

func TestLoad_Success_Dev(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAJADU_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/bm_dev")
	t.Setenv("MAJADU_DB_SCHEMA", "bm_dev")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "dev" {
		t.Errorf("expected Env 'dev', got %q", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected Port '8080', got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost:5432/bm_dev" {
		t.Errorf("expected DatabaseURL 'postgres://localhost:5432/bm_dev', got %q", cfg.DatabaseURL)
	}
	if cfg.DatabaseSchema != "bm_dev" {
		t.Errorf("expected DatabaseSchema 'bm_dev', got %q", cfg.DatabaseSchema)
	}
	if cfg.RateLimitPerMin != 120 {
		t.Errorf("expected RateLimitPerMin 120, got %d", cfg.RateLimitPerMin)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected LogFormat 'text', got %q", cfg.LogFormat)
	}
	if cfg.LogFile != "" {
		t.Errorf("expected LogFile '', got %q", cfg.LogFile)
	}
	if cfg.LogMaxSize != 100 {
		t.Errorf("expected LogMaxSize 100, got %d", cfg.LogMaxSize)
	}
	if cfg.LogMaxAge != 7 {
		t.Errorf("expected LogMaxAge 7, got %d", cfg.LogMaxAge)
	}
	if cfg.LogMaxBackups != 7 {
		t.Errorf("expected LogMaxBackups 7, got %d", cfg.LogMaxBackups)
	}
	if !cfg.LogCompress {
		t.Errorf("expected LogCompress true, got false")
	}
}

func TestLoad_Success_Prod(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAJADU_ENV", "prod")
	t.Setenv("PORT", "9000")
	t.Setenv("DATABASE_URL", "postgres://user:pass@prod-db:5432/bm")
	t.Setenv("MAJADU_DB_SCHEMA", "bm")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://majadu.vercel.app,https://admin.vercel.app")
	t.Setenv("PUBLIC_BASE_URL", "https://api.qouver.com/majadu")
	t.Setenv("MAJADU_RATE_LIMIT_PER_MIN", "60")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_FILE", "/logs/app.log")
	t.Setenv("LOG_MAX_SIZE", "50")
	t.Setenv("LOG_MAX_AGE", "14")
	t.Setenv("LOG_MAX_BACKUPS", "10")
	t.Setenv("LOG_COMPRESS", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "prod" {
		t.Errorf("expected Env 'prod', got %q", cfg.Env)
	}
	if cfg.Port != "9000" {
		t.Errorf("expected Port '9000', got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://user:pass@prod-db:5432/bm" {
		t.Errorf("expected DatabaseURL 'postgres://user:pass@prod-db:5432/bm', got %q", cfg.DatabaseURL)
	}
	if cfg.DatabaseSchema != "bm" {
		t.Errorf("expected DatabaseSchema 'bm', got %q", cfg.DatabaseSchema)
	}
	if cfg.BaseURL != "https://api.qouver.com/majadu" {
		t.Errorf("expected BaseURL 'https://api.qouver.com/majadu', got %q", cfg.BaseURL)
	}
	expectedOrigins := []string{"https://majadu.vercel.app", "https://admin.vercel.app"}
	if !reflect.DeepEqual(cfg.AllowedOrigins, expectedOrigins) {
		t.Errorf("expected AllowedOrigins %#v, got %#v", expectedOrigins, cfg.AllowedOrigins)
	}
	if cfg.RateLimitPerMin != 60 {
		t.Errorf("expected RateLimitPerMin 60, got %d", cfg.RateLimitPerMin)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("expected LogLevel 'warn', got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat 'json', got %q", cfg.LogFormat)
	}
	if cfg.LogFile != "/logs/app.log" {
		t.Errorf("expected LogFile '/logs/app.log', got %q", cfg.LogFile)
	}
	if cfg.LogMaxSize != 50 {
		t.Errorf("expected LogMaxSize 50, got %d", cfg.LogMaxSize)
	}
	if cfg.LogMaxAge != 14 {
		t.Errorf("expected LogMaxAge 14, got %d", cfg.LogMaxAge)
	}
	if cfg.LogMaxBackups != 10 {
		t.Errorf("expected LogMaxBackups 10, got %d", cfg.LogMaxBackups)
	}
	if cfg.LogCompress {
		t.Errorf("expected LogCompress false, got true")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAJADU_ENV", "dev")
	t.Setenv("MAJADU_DB_SCHEMA", "bm_dev")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing, got nil")
	}
}

func TestLoad_MissingDatabaseSchema(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAJADU_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/bm_dev")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when MAJADU_DB_SCHEMA is missing, got nil")
	}
}

func TestLoad_InvalidRateLimitPerMin(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAJADU_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/bm_dev")
	t.Setenv("MAJADU_DB_SCHEMA", "bm_dev")
	t.Setenv("MAJADU_RATE_LIMIT_PER_MIN", "invalid_number")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for non-integer MAJADU_RATE_LIMIT_PER_MIN, got nil")
	}
}
