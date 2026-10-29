// Command api runs the Sawyun Ecommerce HTTP API (and serves the built SPA in production).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/config"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/jobs"
	"github.com/Khant26/business-web-app/backend/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	time.Local = time.UTC
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	if !cfg.IsDev() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		return err
	}
	slog.Info("migrations applied")
	if err := auth.EnsureAdmin(ctx, pool, cfg.AdminEmail); err != nil {
		return err
	}
	for _, d := range []string{filepath.Join(cfg.UploadDir, "public"), filepath.Join(cfg.UploadDir, "private", "slips")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	var mailer auth.Mailer = auth.LogMailer{}
	if cfg.SMTPHost != "" {
		smtpMailer := &auth.SMTPMailer{Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUsername, Password: cfg.SMTPPassword, From: cfg.SMTPFrom}
		mailer = smtpMailer
		if cfg.IsDev() {
			mailer = auth.DevMailer{SMTP: smtpMailer}
		}
	} else {
		slog.Warn("OTP emails are logged, not sent (SMTP_HOST empty)")
	}

	jobs.Start(ctx, pool, cfg.SessionIdle)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(ctx, cfg, pool, mailer),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}
	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}
	slog.Info("shutting down")
	sctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(sctx)
}
