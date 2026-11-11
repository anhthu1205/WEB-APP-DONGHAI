package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

const userCols = `id, username, full_name, password_hash, role, is_active, token_version, last_login_at, created_at, updated_at`

func scanUser(r pgx.Row) (*domain.User, error) {
	var u domain.User
	err := r.Scan(&u.ID, &u.Username, &u.FullName, &u.PasswordHash, &u.Role, &u.IsActive,
		&u.TokenVersion, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

// CountAdmins đếm số tài khoản admin.
func (s *Store) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := s.q.QueryRow(ctx, `SELECT count(*) FROM users WHERE role = 'admin'`).Scan(&n)
	return n, mapErr(err)
}

// CreateUser thêm tài khoản.
func (s *Store) CreateUser(ctx context.Context, username, fullName, hash, role string) (*domain.User, error) {
	return scanUser(s.q.QueryRow(ctx, `INSERT INTO users (username, full_name, password_hash, role)
		VALUES ($1, $2, $3, $4) RETURNING `+userCols, username, fullName, hash, role))
}

// GetUserByID lấy user theo id.
func (s *Store) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return scanUser(s.q.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id = $1`, id))
}

// GetUserByUsername lấy user theo username (lowercase).
func (s *Store) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return scanUser(s.q.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE username = $1`, username))
}

// ListUsers liệt kê tài khoản (admin trước).
func (s *Store) ListUsers(ctx context.Context) ([]*domain.User, error) {
	rows, err := s.q.Query(ctx, `SELECT `+userCols+` FROM users ORDER BY role = 'admin' DESC, id`)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(r pgx.Rows) (*domain.User, error) { return scanUser(r) })
}

// UpdateUser cập nhật full_name / is_active; bumpTV tăng token_version.
func (s *Store) UpdateUser(ctx context.Context, id int64, fullName *string, isActive *bool, bumpTV bool) (*domain.User, error) {
	return scanUser(s.q.QueryRow(ctx, `UPDATE users SET
		full_name = COALESCE($2, full_name),
		is_active = COALESCE($3, is_active),
		token_version = token_version + CASE WHEN $4 THEN 1 ELSE 0 END
		WHERE id = $1 RETURNING `+userCols, id, fullName, isActive, bumpTV))
}

// SetPassword đổi hash mật khẩu và tăng token_version.
func (s *Store) SetPassword(ctx context.Context, id int64, hash string) (*domain.User, error) {
	return scanUser(s.q.QueryRow(ctx, `UPDATE users SET password_hash = $2, token_version = token_version + 1
		WHERE id = $1 RETURNING `+userCols, id, hash))
}

// TouchLogin ghi thời điểm đăng nhập.
func (s *Store) TouchLogin(ctx context.Context, id int64) error {
	_, err := s.q.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, id)
	return mapErr(err)
}
