package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// RequestIDFrom trả request id (chi middleware.RequestID).
func RequestIDFrom(ctx context.Context) string { return chimw.GetReqID(ctx) }

// SecureHeaders đặt header bảo mật trên mọi response.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// CORS chỉ bật khi có danh sách origin (dev); rỗng → không thêm header nào.
func CORS(origins []string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		if len(allowed) == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Max-Age", "600")
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Logger ghi log request dạng slog JSON (có request_id, IP client). Không log body/header nhạy cảm.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			path := r.URL.Path
			if strings.HasPrefix(path, "/assets/") {
				return // bớt log file tĩnh
			}
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			lvl := slog.LevelInfo
			if status >= 500 {
				lvl = slog.LevelError
			}
			// chỉ log path (không query) để không lộ dữ liệu
			log.Log(r.Context(), lvl, "request",
				"request_id", RequestIDFrom(r.Context()),
				"method", r.Method, "path", path, "status", status,
				"bytes", ww.BytesWritten(), "duration_ms", time.Since(start).Milliseconds(),
				"ip", ClientIP(r))
		})
	}
}

// Recoverer bắt panic → 500 JSON, log stack kèm request_id.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if rec == http.ErrAbortHandler {
						panic(rec)
					}
					log.Error("panic", "request_id", RequestIDFrom(r.Context()), "panic", rec, "stack", string(debug.Stack()))
					WriteError(w, http.StatusInternalServerError, "INTERNAL", "Đã có lỗi xảy ra")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
