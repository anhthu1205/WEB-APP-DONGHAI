// Command server: config → DB → migrate → seed admin/demo → router → ListenAndServe + graceful shutdown.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"donghai/internal/auth"
	"donghai/internal/config"
	"donghai/internal/db"
	"donghai/internal/handler"
	"donghai/internal/repository"
	"donghai/internal/service"
	"donghai/migrations"
	"donghai/seed"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server dừng do lỗi", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error(err.Error())
		os.Exit(2)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns, log)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.MigrateOnStart {
		applied, err := db.Migrate(ctx, pool, migrations.FS, log)
		if err != nil {
			return err
		}
		log.Info("migration xong", "applied", applied)
	}

	svc := service.New(repository.New(pool), auth.NewIssuer(cfg.JWTSecret, cfg.JWTTTL, nil), nil)

	created, err := svc.BootstrapAdmin(ctx, cfg.AdminUsername, cfg.AdminPassword)
	if err != nil {
		return err
	}
	if created {
		log.Info("đã tạo tài khoản admin ban đầu", "username", cfg.AdminUsername)
	}

	if cfg.SeedDemo {
		seeded, err := db.SeedDemo(ctx, pool, seed.DemoSQL)
		if err != nil {
			return err
		}
		log.Info("seed demo", "seeded", seeded)
	}

	staticDir := cfg.StaticDir
	if staticDir != "" {
		if fi, err := os.Stat(staticDir); err != nil || !fi.IsDir() {
			log.Warn("STATIC_DIR không tồn tại, chỉ phục vụ API", "static_dir", staticDir)
			staticDir = ""
		}
	}

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: handler.NewRouter(handler.Options{
			Service: svc, Logger: log, StaticDir: staticDir,
			AllowedOrigins: cfg.AllowedOrigins, Limits: handler.DefaultLimits(),
		}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server đang lắng nghe", "port", cfg.Port)
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
	log.Info("đang tắt server…")
	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shCtx)
}
