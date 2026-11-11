// Package testutil: tạo DB test (reset schema + migrate) và server httptest.
package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"donghai/internal/auth"
	"donghai/internal/db"
	"donghai/internal/handler"
	"donghai/internal/repository"
	"donghai/internal/service"
	"donghai/migrations"
)

// TestSecret là JWT secret dùng trong test.
const TestSecret = "test-secret-0123456789-abcdefghijklmnop"

// FixedNow: 2026-11-11 10:00 giờ VN (khớp ví dụ trong tài liệu).
func FixedNow() time.Time {
	return time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)
}

// Pool trả pool tới TEST_DATABASE_URL, dùng một schema riêng (search_path) cho mỗi test
// để các package test chạy song song không giẫm lên nhau. Bỏ qua test nếu thiếu env.
// Không chạy migration.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("bỏ qua: chưa đặt TEST_DATABASE_URL")
	}
	auth.BcryptCost = bcrypt.MinCost
	ctx := context.Background()
	schema := fmt.Sprintf("t_%d_%d", time.Now().UnixNano(), rand.Intn(1_000_000))
	admin, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("kết nối DB test: %v", err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("tạo schema: %v", err)
	}
	_ = admin.Close(ctx)

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("kết nối DB test: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		if c, err := pgx.Connect(context.Background(), url); err == nil {
			_, _ = c.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
			_ = c.Close(context.Background())
		}
	})
	return pool
}

// MigratedPool = Pool + chạy migration.
func MigratedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := Pool(t)
	if _, err := db.Migrate(context.Background(), pool, migrations.FS, nil); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

// NewService tạo Service với thời gian cố định.
func NewService(pool *pgxpool.Pool, now func() time.Time) *service.Service {
	return service.New(repository.New(pool), auth.NewIssuer(TestSecret, 12*time.Hour, now), now)
}

// NewServer tạo httptest server với router thật.
func NewServer(t *testing.T, svc *service.Service, limits handler.Limits, staticDir string) *httptest.Server {
	t.Helper()
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := httptest.NewServer(handler.NewRouter(handler.Options{Service: svc, Logger: log, StaticDir: staticDir, Limits: limits}))
	t.Cleanup(srv.Close)
	return srv
}

// Resp là phản hồi HTTP đã đọc body.
type Resp struct {
	Status int
	Header http.Header
	Body   []byte
}

// JSON giải mã body vào v.
func (r *Resp) JSON(t *testing.T, v any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, v); err != nil {
		t.Fatalf("giải mã JSON: %v; body=%s", err, r.Body)
	}
}

// Map giải mã body thành map.
func (r *Resp) Map(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	r.JSON(t, &m)
	return m
}

// Do gửi request; body là []byte (gửi nguyên) hoặc giá trị bất kỳ (JSON hóa).
func Do(t *testing.T, method, url, token string, body any, headers ...string) *Resp {
	t.Helper()
	var rd io.Reader
	ct := ""
	switch b := body.(type) {
	case nil:
	case []byte:
		rd = bytes.NewReader(b)
		ct = "application/json"
	default:
		buf, _ := json.Marshal(b)
		rd = bytes.NewReader(buf)
		ct = "application/json"
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		t.Fatal(err)
	}
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	return send(t, req)
}

// DoRaw gửi request đã dựng sẵn.
func DoRaw(t *testing.T, req *http.Request) *Resp {
	t.Helper()
	return send(t, req)
}

func send(t *testing.T, req *http.Request) *Resp {
	t.Helper()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return &Resp{Status: res.StatusCode, Header: res.Header, Body: b}
}
