package service

import (
	"context"
	"strings"

	"donghai/internal/auth"
	"donghai/internal/domain"
)

// CreateUserRequest là body POST /api/admin/users (luôn tạo role user).
type CreateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

// UpdateUserRequest là body PATCH /api/admin/users/{id} (không có role).
type UpdateUserRequest struct {
	FullName *string `json:"full_name"`
	IsActive *bool   `json:"is_active"`
}

// ResetPasswordRequest là body POST /api/admin/users/{id}/reset-password.
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// ListUsers liệt kê tài khoản.
func (s *Service) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.store.ListUsers(ctx)
}

// CreateUser tạo tài khoản role user.
func (s *Service) CreateUser(ctx context.Context, in CreateUserRequest) (*domain.User, error) {
	username := strings.ToLower(strings.TrimSpace(in.Username))
	fullName := strings.TrimSpace(in.FullName)
	var v domain.Validator
	v.Check(usernameRe.MatchString(username), "username", "Tên đăng nhập 3–32 ký tự, chỉ gồm chữ thường, số, dấu _ và .")
	v.Check(runeLen(fullName) <= 100, "full_name", "Họ tên tối đa 100 ký tự")
	validatePassword(&v, "password", in.Password)
	if err := v.Err(); err != nil {
		return nil, err
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	return s.store.CreateUser(ctx, username, fullName, hash, domain.RoleUser)
}

// UpdateUser sửa họ tên / khóa-mở khóa. Không cho tự khóa mình hoặc khóa admin.
func (s *Service) UpdateUser(ctx context.Context, me *domain.User, id int64, in UpdateUserRequest) (*domain.User, error) {
	target, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy tài khoản")
	}
	var v domain.Validator
	var fullName *string
	if in.FullName != nil {
		f := strings.TrimSpace(*in.FullName)
		v.Check(runeLen(f) <= 100, "full_name", "Họ tên tối đa 100 ký tự")
		fullName = &f
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	bump := false
	if in.IsActive != nil && *in.IsActive != target.IsActive {
		if !*in.IsActive {
			if target.ID == me.ID {
				return nil, domain.BadRequest("Bạn không thể tự khóa tài khoản của mình")
			}
			if target.Role == domain.RoleAdmin {
				return nil, domain.BadRequest("Không thể khóa tài khoản admin")
			}
		}
		bump = true
	}
	return s.store.UpdateUser(ctx, id, fullName, in.IsActive, bump)
}

// ResetPassword đặt lại mật khẩu cho user (tăng token_version).
func (s *Service) ResetPassword(ctx context.Context, id int64, in ResetPasswordRequest) error {
	var v domain.Validator
	validatePassword(&v, "new_password", in.NewPassword)
	if err := v.Err(); err != nil {
		return err
	}
	if _, err := s.store.GetUserByID(ctx, id); err != nil {
		return notFound(err, "Không tìm thấy tài khoản")
	}
	hash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	_, err = s.store.SetPassword(ctx, id, hash)
	return err
}
