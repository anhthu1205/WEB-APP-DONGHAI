// Package config đọc và kiểm tra biến môi trường, dừng sớm khi thiếu cấu hình bắt buộc.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Config là cấu hình bất biến của ứng dụng.
type Config struct {
	DatabaseURL    string
	JWTSecret      string
	JWTTTL         time.Duration
	AdminUsername  string
	AdminPassword  string
	Port           string
	StaticDir      string
	MigrateOnStart bool
	SeedDemo       bool
	AllowedOrigins []string
	DBMaxConns     int32
	LogLevel       slog.Level
}

// MinJWTSecretLen là độ dài tối thiểu của JWT_SECRET.
const MinJWTSecretLen = 32

// Load đọc cấu hình từ hàm getenv (thường là os.Getenv). Trả lỗi tổng hợp nếu sai.
func Load(getenv func(string) string) (*Config, error) {
	var errs []string
	get := func(k string) string { return strings.TrimSpace(getenv(k)) }

	c := &Config{
		DatabaseURL:   get("DATABASE_URL"),
		JWTSecret:     get("JWT_SECRET"),
		AdminUsername: strings.ToLower(get("ADMIN_USERNAME")),
		AdminPassword: getenv("ADMIN_PASSWORD"),
		Port:          get("PORT"),
		StaticDir:     get("STATIC_DIR"),
	}
	if c.DatabaseURL == "" {
		errs = append(errs, "thiếu DATABASE_URL")
	}
	if c.JWTSecret == "" {
		errs = append(errs, "thiếu JWT_SECRET")
	} else if len(c.JWTSecret) < MinJWTSecretLen {
		errs = append(errs, fmt.Sprintf("JWT_SECRET phải dài ít nhất %d ký tự", MinJWTSecretLen))
	}
	if c.Port == "" {
		c.Port = "8080"
	}

	ttl, err := intEnv(get("JWT_TTL_HOURS"), 12)
	if err != nil || ttl < 1 || ttl > 24*30 {
		errs = append(errs, "JWT_TTL_HOURS không hợp lệ (1..720)")
	}
	c.JWTTTL = time.Duration(ttl) * time.Hour

	c.MigrateOnStart, err = boolEnv(get("MIGRATE_ON_START"), true)
	if err != nil {
		errs = append(errs, "MIGRATE_ON_START không hợp lệ")
	}
	c.SeedDemo, err = boolEnv(get("SEED_DEMO"), false)
	if err != nil {
		errs = append(errs, "SEED_DEMO không hợp lệ")
	}

	maxConns, err := intEnv(get("DB_MAX_CONNS"), 5)
	if err != nil || maxConns < 1 || maxConns > 100 {
		errs = append(errs, "DB_MAX_CONNS không hợp lệ (1..100)")
	}
	c.DBMaxConns = int32(maxConns)

	for _, o := range strings.Split(get("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.AllowedOrigins = append(c.AllowedOrigins, strings.TrimRight(o, "/"))
		}
	}

	switch strings.ToLower(get("LOG_LEVEL")) {
	case "", "info":
		c.LogLevel = slog.LevelInfo
	case "debug":
		c.LogLevel = slog.LevelDebug
	case "warn", "warning":
		c.LogLevel = slog.LevelWarn
	case "error":
		c.LogLevel = slog.LevelError
	default:
		errs = append(errs, "LOG_LEVEL không hợp lệ (debug|info|warn|error)")
	}

	if len(errs) > 0 {
		return nil, errors.New("cấu hình không hợp lệ: " + strings.Join(errs, "; "))
	}
	return c, nil
}

func intEnv(v string, def int) (int, error) {
	if v == "" {
		return def, nil
	}
	return strconv.Atoi(v)
}

func boolEnv(v string, def bool) (bool, error) {
	if v == "" {
		return def, nil
	}
	return strconv.ParseBool(v)
}
