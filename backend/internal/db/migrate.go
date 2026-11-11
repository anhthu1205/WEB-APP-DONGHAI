package db

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrateLockID = 727001

const createMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
  version    text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
)`

// Migrate chạy các file *.sql trong fsys theo thứ tự tên, bỏ qua version đã có.
// Toàn bộ quá trình giữ pg_advisory_lock trên MỘT kết nối riêng.
// Trả danh sách version vừa áp dụng.
func Migrate(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, log *slog.Logger) ([]string, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy kết nối migrate: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrateLockID); err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}
	defer func() {
		// unlock trên chính kết nối đã lock
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", migrateLockID)
	}()

	if _, err := conn.Exec(ctx, createMigrationsTable); err != nil {
		return nil, fmt.Errorf("tạo schema_migrations: %w", err)
	}

	applied := map[string]bool{}
	rows, err := conn.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("đọc schema_migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return nil, err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	var done []string
	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return done, err
		}
		err = pgx.BeginFunc(ctx, conn.Conn(), func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, string(body)); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", version)
			return err
		})
		if err != nil {
			return done, fmt.Errorf("migration %s: %w", version, err)
		}
		if log != nil {
			log.Info("đã áp dụng migration", "version", version)
		}
		done = append(done, version)
	}
	return done, nil
}

// SeedDemo chạy seed nếu bảng products rỗng. Trả true nếu đã seed.
func SeedDemo(ctx context.Context, pool *pgxpool.Pool, sql string) (bool, error) {
	seeded := false
	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		// khóa giao dịch để 2 instance không seed song song
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", migrateLockID+1); err != nil {
			return err
		}
		var n int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM products").Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return nil
		}
		if _, err := tx.Exec(ctx, sql); err != nil {
			return err
		}
		seeded = true
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("seed demo: %w", err)
	}
	return seeded, nil
}
