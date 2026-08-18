package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"majadu-api/internal/config"
	"majadu-api/internal/db"
	"majadu-api/internal/handler"
	"majadu-api/internal/logger"
	"majadu-api/internal/middleware"
	"majadu-api/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Fail-fast: prod harus punya config lengkap.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	// Logger: stdout default; kalau LOG_FILE di-set → rolling log file.
	logger, cleanup := logger.NewLogger(logger.Options{
		Level:      cfg.LogLevel,
		Format:     cfg.LogFormat,
		Filename:   cfg.LogFile,
		MaxSize:    cfg.LogMaxSize,
		MaxAge:     cfg.LogMaxAge,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
	})
	if cleanup != nil {
		defer func() { _ = cleanup() }()
	}
	slog.SetDefault(logger)

	// Context dibatalkan saat SIGINT/SIGTERM; dipakai juga untuk
	// memberhentikan janitor rate limiter saat shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL, cfg.DatabaseSchema, logger)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connected")

	mux := http.NewServeMux()
	h := registerRoutes(mux, logger, cfg, pool, ctx)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown.
	<-ctx.Done()
	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}

func registerRoutes(mux *http.ServeMux, logger *slog.Logger, cfg config.Config, pool *pgxpool.Pool, ctx context.Context) http.Handler {
	health := &handler.HealthHandler{Pool: pool}
	mux.Handle("GET /healthz", http.HandlerFunc(health.Healthz))
	mux.Handle("GET /readyz", http.HandlerFunc(health.Ready))
	mux.Handle("GET /version", http.HandlerFunc(health.Version))

	sessions := &handler.SessionHandler{Store: store.NewSessionStore(pool, cfg.DatabaseSchema), Logger: logger, BaseURL: cfg.BaseURL}
	mux.Handle("GET /sessions", http.HandlerFunc(sessions.List))
	mux.Handle("POST /sessions", http.HandlerFunc(sessions.Create))
	mux.Handle("GET /sessions/{id}", http.HandlerFunc(sessions.Get))
	mux.Handle("PUT /sessions/{id}", http.HandlerFunc(sessions.Put))
	mux.Handle("PATCH /sessions/{id}", http.HandlerFunc(sessions.Patch))
	mux.Handle("DELETE /sessions/{id}", http.HandlerFunc(sessions.Delete))
	mux.Handle("POST /sessions/{id}/lock", http.HandlerFunc(sessions.Lock))
	mux.Handle("POST /sessions/{id}/unlock", http.HandlerFunc(sessions.Unlock))

	players := &handler.PlayerHandler{Store: store.NewPlayerStore(pool), Logger: logger}
	mux.Handle("GET /players", http.HandlerFunc(players.List))
	mux.Handle("POST /players", http.HandlerFunc(players.Register))
	mux.Handle("GET /players/{name}/stats", http.HandlerFunc(players.Stats))

	tournaments := &handler.TournamentHandler{Store: store.NewTournamentStore(pool, cfg.DatabaseSchema), Logger: logger, BaseURL: cfg.BaseURL}
	mux.Handle("GET /tournaments", http.HandlerFunc(tournaments.List))
	mux.Handle("POST /tournaments", http.HandlerFunc(tournaments.Create))
	mux.Handle("GET /tournaments/{id}", http.HandlerFunc(tournaments.Get))
	mux.Handle("PUT /tournaments/{id}", http.HandlerFunc(tournaments.Put))
	mux.Handle("PATCH /tournaments/{id}", http.HandlerFunc(tournaments.Patch))

	// Middleware chain: recover (luar) → request-id → logging → CORS → rate limit → mux.
	var h http.Handler = mux
	h = middleware.RateLimit(ctx, cfg.RateLimitPerMin, logger)(h)
	h = middleware.CORS(cfg.AllowedOrigins)(h)
	h = middleware.Logging(logger)(h)
	h = middleware.RequestID(h)
	h = middleware.Recover(logger)(h)
	return h
}
