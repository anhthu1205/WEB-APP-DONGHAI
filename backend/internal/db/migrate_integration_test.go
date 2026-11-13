//go:build integration

package db_test

import (
	"context"
	"testing"

	"donghai/internal/db"
	"donghai/internal/testutil"
	"donghai/migrations"
	"donghai/seed"
)

func TestMigrateTwiceAndSeed(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()

	applied, err := db.Migrate(ctx, pool, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 5 {
		t.Fatalf("lần 1 phải áp dụng 5 migration, được %v", applied)
	}
	applied, err = db.Migrate(ctx, pool, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 0 {
		t.Fatalf("lần 2 không được chạy lại, được %v", applied)
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&n); err != nil || n != 5 {
		t.Fatalf("schema_migrations phải có 5 dòng, được %d (%v)", n, err)
	}
	// RLS bật trên mọi bảng
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_class c JOIN pg_namespace ns ON ns.oid = c.relnamespace
		WHERE ns.nspname = current_schema() AND c.relkind = 'r' AND NOT c.relrowsecurity`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("còn %d bảng chưa bật RLS (%v)", n, err)
	}

	seeded, err := db.SeedDemo(ctx, pool, seed.DemoSQL)
	if err != nil || !seeded {
		t.Fatalf("seed lần 1: %v %v", seeded, err)
	}
	seeded, err = db.SeedDemo(ctx, pool, seed.DemoSQL)
	if err != nil || seeded {
		t.Fatalf("seed lần 2 phải bỏ qua: %v %v", seeded, err)
	}
	counts := map[string]int{
		"SELECT count(*) FROM products":                      8,
		"SELECT count(*) FROM products WHERE NOT is_visible": 1,
		"SELECT count(*) FROM promotions":                    3,
		"SELECT count(*) FROM customers":                     5,
		"SELECT count(*) FROM orders":                        8,
	}
	for q, want := range counts {
		if err := pool.QueryRow(ctx, q).Scan(&n); err != nil || n != want {
			t.Errorf("%s = %d (%v); muốn %d", q, n, err, want)
		}
	}
}
