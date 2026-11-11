package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"donghai/internal/auth"
	"donghai/internal/domain"
)

const msgBadLogin = "Sai tên đăng nhập hoặc mật khẩu"

var usernameRe = regexp.MustCompile(`^[a-z0-9_.]{3,32}$`)

// LoginRequest là body POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest là body POST /api/auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Session là kết quả đăng nhập.
type Session struct {
	Token     string
	ExpiresAt time.Time
	User      *domain.User
}

// Login kiểm tra mật khẩu, trả token. Mọi lỗi đăng nhập đều là 401 message chung.
func (s *Service) Login(ctx context.Context, in LoginRequest) (*Session, error) {
	username := strings.ToLower(strings.TrimSpace(in.Username))
	if username == "" || in.Password == "" || len(in.Password) > auth.MaxPasswordLen {
		return nil, domain.Unauthorized(msgBadLogin)
	}
	u, err := s.store.GetUserByUsername(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		auth.BurnPasswordCheck(in.Password)
		return nil, domain.Unauthorized(msgBadLogin)
	}
	if err != nil {
		return nil, err
	}
	if !auth.CheckPassword(u.PasswordHash, in.Password) || !u.IsActive {
		return nil, domain.Unauthorized(msgBadLogin)
	}
	tok, exp, err := s.jwt.Sign(u.ID, u.Role, u.TokenVersion)
	if err != nil {
		return nil, err
	}
	if err := s.store.TouchLogin(ctx, u.ID); err != nil {
		return nil, err
	}
	return &Session{Token: tok, ExpiresAt: exp, User: u}, nil
}

// Authenticate kiểm tra JWT, nạp user từ DB; từ chối nếu bị khóa hoặc token_version lệch.
func (s *Service) Authenticate(ctx context.Context, token string) (*domain.User, error) {
	c, err := s.jwt.Parse(token)
	if err != nil {
		return nil, domain.Unauthorized("Phiên đăng nhập không hợp lệ hoặc đã hết hạn")
	}
	id, _ := c.UserID()
	u, err := s.store.GetUserByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.Unauthorized("Phiên đăng nhập không hợp lệ hoặc đã hết hạn")
	}
	if err != nil {
		return nil, err
	}
	if !u.IsActive || u.TokenVersion != c.TV {
		return nil, domain.Unauthorized("Phiên đăng nhập không hợp lệ hoặc đã hết hạn")
	}
	return u, nil
}

func validatePassword(v *domain.Validator, field, pw string) {
	if len(pw) < auth.MinPasswordLen || len(pw) > auth.MaxPasswordLen {
		v.Add(field, "Mật khẩu phải từ 8 đến 72 ký tự")
	}
}

// ChangePassword đổi mật khẩu của chính mình, trả token mới (token cũ mất hiệu lực).
func (s *Service) ChangePassword(ctx context.Context, me *domain.User, in ChangePasswordRequest) (*Session, error) {
	var v domain.Validator
	if !auth.CheckPassword(me.PasswordHash, in.CurrentPassword) {
		v.Add("current_password", "Mật khẩu hiện tại không đúng")
	}
	validatePassword(&v, "new_password", in.NewPassword)
	if in.NewPassword != "" && in.NewPassword == in.CurrentPassword {
		v.Add("new_password", "Mật khẩu mới phải khác mật khẩu hiện tại")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	hash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		return nil, err
	}
	u, err := s.store.SetPassword(ctx, me.ID, hash)
	if err != nil {
		return nil, err
	}
	tok, exp, err := s.jwt.Sign(u.ID, u.Role, u.TokenVersion)
	if err != nil {
		return nil, err
	}
	return &Session{Token: tok, ExpiresAt: exp, User: u}, nil
}
