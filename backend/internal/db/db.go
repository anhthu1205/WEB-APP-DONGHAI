// Package db: kết nối pgxpool và chạy migration/seed.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect tạo pool và ping, thử lại tối đa 5 lần (backoff 1s,2s,4s,8s).
func Connect(ctx context.Context, url string, maxConns int32, log *slog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		// không in url (có mật khẩu)
		return nil, fmt.Errorf("DATABASE_URL không hợp lệ")
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	var lastErr error
	backoff := time.Second
	for attempt := 1; attempt <= 5; attempt++ {
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pctx)
			cancel()
			if err == nil {
				return pool, nil
			}
			pool.Close()
		}
		lastErr = err
		if log != nil {
			log.Warn("kết nối DB thất bại, thử lại", "attempt", attempt, "error", err.Error())
		}
		if attempt == 5 {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return nil, fmt.Errorf("không kết nối được DB sau 5 lần: %w", lastErr)
}
