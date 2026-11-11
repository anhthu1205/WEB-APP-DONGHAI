package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"donghai/internal/auth"
	"donghai/internal/domain"
)

// ErrAdminEnvMissing: DB chưa có admin và thiếu ADMIN_USERNAME/ADMIN_PASSWORD.
var ErrAdminEnvMissing = errors.New("chưa có tài khoản admin: cần đặt ADMIN_USERNAME và ADMIN_PASSWORD")

// BootstrapAdmin tạo admin từ env khi DB chưa có admin. Đã có admin → không làm gì
// (không ghi đè mật khẩu). Trả true nếu vừa tạo.
func (s *Service) BootstrapAdmin(ctx context.Context, username, password string) (bool, error) {
	n, err := s.store.CountAdmins(ctx)
	if err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || password == "" {
		return false, ErrAdminEnvMissing
	}
	if !usernameRe.MatchString(username) {
		return false, fmt.Errorf("ADMIN_USERNAME không hợp lệ (3–32 ký tự a-z 0-9 _ .)")
	}
	if len(password) < auth.MinPasswordLen || len(password) > auth.MaxPasswordLen {
		return false, fmt.Errorf("ADMIN_PASSWORD phải từ 8 đến 72 ký tự")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return false, err
	}
	if _, err := s.store.CreateUser(ctx, username, "Quản trị viên", hash, domain.RoleAdmin); err != nil {
		return false, fmt.Errorf("tạo admin: %w", err)
	}
	return true, nil
}
