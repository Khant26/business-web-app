// Package jobs runs periodic background maintenance.
package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/orders"
)

// Start launches the background jobs; they stop when ctx is cancelled.
func Start(ctx context.Context, pool *pgxpool.Pool, sessionIdle time.Duration) {
	go every(ctx, time.Minute, func(ctx context.Context) {
		n, err := orders.ExpireDue(ctx, pool)
		if err != nil {
			slog.Error("expire orders", "err", err)
		} else if n > 0 {
			slog.Info("expired unpaid orders", "count", n)
		}
	})
	go every(ctx, 30*time.Minute, func(ctx context.Context) {
		if err := Purge(ctx, pool, sessionIdle); err != nil {
			slog.Error("purge", "err", err)
		}
	})
}

// Purge deletes OTP codes older than a day and expired/idle sessions.
func Purge(ctx context.Context, pool *pgxpool.Pool, sessionIdle time.Duration) error {
	if _, err := pool.Exec(ctx, `DELETE FROM otp_codes WHERE created_at < now() - interval '1 day'`); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now() OR last_seen_at < now() - $1::interval`, sessionIdle)
	return err
}

func every(ctx context.Context, d time.Duration, fn func(context.Context)) {
	run := func() {
		c, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		fn(c)
	}
	run()
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			run()
		}
	}
}
