# Lệnh tiện dụng. Biến môi trường đọc từ .env nếu có (không commit .env).
SHELL := /bin/bash
export GOTOOLCHAIN := local
-include .env
export

PG_BIN ?= /usr/lib/postgresql/16/bin
PG_DIR ?= /var/tmp/donghai-pg
PG_PORT ?= 54329
TEST_DATABASE_URL ?= postgres://postgres@localhost:$(PG_PORT)/donghai_test?sslmode=disable

.PHONY: help pg-start pg-stop dev-api dev-web build build-web build-api test test-int check docker

help: ## Liệt kê lệnh
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-12s %s\n", $$1, $$2}'

pg-start: ## Khởi động Postgres 16 local (chạy bằng root; dữ liệu ở PG_DIR, cổng PG_PORT)
	mkdir -p $(PG_DIR) && chown postgres $(PG_DIR)
	[ -f $(PG_DIR)/PG_VERSION ] || su postgres -s /bin/sh -c "$(PG_BIN)/initdb -D $(PG_DIR) -U postgres -A trust >/dev/null"
	su postgres -s /bin/sh -c "$(PG_BIN)/pg_ctl -D $(PG_DIR) -o '-p $(PG_PORT) -k /tmp' -l $(PG_DIR)/log start -w"
	psql -h /tmp -p $(PG_PORT) -U postgres -c 'CREATE DATABASE donghai_test' || true
	psql -h /tmp -p $(PG_PORT) -U postgres -c 'CREATE DATABASE donghai_dev' || true

pg-stop: ## Dừng Postgres local
	su postgres -s /bin/sh -c "$(PG_BIN)/pg_ctl -D $(PG_DIR) stop"

dev-api: ## Chạy backend (đọc .env)
	cd backend && go run ./cmd/server

dev-web: ## Chạy Vite dev server :5173 (proxy /api → :8080)
	cd frontend && npm run dev

build-web: ## Build frontend → frontend/dist
	cd frontend && npm ci && npm run build

build-api: ## Build binary backend → backend/bin/server
	cd backend && go build -o bin/server ./cmd/server

build: build-web build-api ## Build cả hai

test: ## Unit test Go (không cần DB)
	cd backend && go test ./...

test-int: ## Unit + tích hợp với Postgres local
	cd backend && TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test -count=1 -tags=integration ./...

check: ## gofmt + vet + test tích hợp + build frontend
	cd backend && test -z "$$(gofmt -l .)" && go vet ./... && go vet -tags=integration ./...
	$(MAKE) test-int
	cd frontend && npm run build

docker: ## Build Docker image
	docker build -t web-app-donghai .
