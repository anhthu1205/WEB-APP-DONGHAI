# --- 1. Build frontend (Vue + Vite) ---
FROM node:22-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# --- 2. Build backend (Go) ---
FROM golang:1.24-alpine AS api
ENV GOTOOLCHAIN=local CGO_ENABLED=0
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- 3. Runtime tối giản (không shell, user nonroot) ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=api /out/server /app/server
COPY --from=web /src/frontend/dist /app/web
ENV STATIC_DIR=/app/web PORT=8080
EXPOSE 8080
USER nonroot
CMD ["/app/server"]
