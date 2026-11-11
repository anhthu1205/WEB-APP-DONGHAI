package config

import (
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadRequired(t *testing.T) {
	cases := []struct {
		name string
		m    map[string]string
		want string
	}{
		{"thiếu DATABASE_URL", map[string]string{"JWT_SECRET": strings.Repeat("x", 32)}, "DATABASE_URL"},
		{"thiếu JWT_SECRET", map[string]string{"DATABASE_URL": "postgres://x"}, "JWT_SECRET"},
		{"JWT_SECRET ngắn", map[string]string{"DATABASE_URL": "postgres://x", "JWT_SECRET": "short"}, "ít nhất 32"},
		{"TTL sai", map[string]string{"DATABASE_URL": "postgres://x", "JWT_SECRET": strings.Repeat("x", 32), "JWT_TTL_HOURS": "abc"}, "JWT_TTL_HOURS"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Load(env(c.m))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("muốn lỗi chứa %q, được %v", c.want, err)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	c, err := Load(env(map[string]string{
		"DATABASE_URL":    "postgres://x",
		"JWT_SECRET":      strings.Repeat("s", 40),
		"ADMIN_USERNAME":  " Admin ",
		"ALLOWED_ORIGINS": "http://localhost:5173/, ,http://a.b",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "8080" || c.JWTTTL != 12*time.Hour || !c.MigrateOnStart || c.SeedDemo || c.DBMaxConns != 5 {
		t.Fatalf("mặc định sai: %+v", c)
	}
	if c.AdminUsername != "admin" {
		t.Fatalf("username phải lowercase: %q", c.AdminUsername)
	}
	if len(c.AllowedOrigins) != 2 || c.AllowedOrigins[0] != "http://localhost:5173" {
		t.Fatalf("origins sai: %v", c.AllowedOrigins)
	}
}
