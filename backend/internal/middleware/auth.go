package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"donghai/internal/domain"
)

type ctxKey int

const userKey ctxKey = 1

// Authenticator xác thực token → user (đã kiểm khóa/token_version).
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*domain.User, error)
}

// UserFromContext trả user hiện tại (nil nếu chưa xác thực).
func UserFromContext(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userKey).(*domain.User)
	return u
}

// WithUser gắn user vào context (dùng trong test).
func WithUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// RequireAuth bắt buộc JWT hợp lệ; role lấy từ DB (không tin claim).
func RequireAuth(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok := bearerToken(r)
			if tok == "" {
				WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Vui lòng đăng nhập")
				return
			}
			u, err := a.Authenticate(r.Context(), tok)
			if err != nil {
				if errors.Is(err, domain.ErrUnauthorized) {
					WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn")
					return
				}
				slog.ErrorContext(r.Context(), "xác thực lỗi", "request_id", RequestIDFrom(r.Context()), "error", err.Error())
				WriteError(w, http.StatusInternalServerError, "INTERNAL", "Đã có lỗi xảy ra")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
		})
	}
}

// RequireAdmin chỉ cho role admin (đặt sau RequireAuth).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Vui lòng đăng nhập")
			return
		}
		if !u.IsAdmin() {
			WriteError(w, http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền thực hiện thao tác này")
			return
		}
		next.ServeHTTP(w, r)
	})
}
