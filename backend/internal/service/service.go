// Package service chứa nghiệp vụ: validate, quyền nghiệp vụ, transaction.
package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"donghai/internal/auth"
	"donghai/internal/domain"
	"donghai/internal/repository"
	"donghai/internal/warranty"
)

// Service gom toàn bộ nghiệp vụ, dùng chung Store.
type Service struct {
	store *repository.Store
	jwt   *auth.Issuer
	now   func() time.Time
}

// New tạo Service. now = nil → time.Now (inject được trong test).
func New(store *repository.Store, issuer *auth.Issuer, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, jwt: issuer, now: now}
}

// Today trả ngày hôm nay (giờ VN) dạng 00:00 UTC.
func (s *Service) Today() time.Time { return warranty.Today(s.now()) }

// Ping kiểm tra DB.
func (s *Service) Ping(ctx context.Context) error { return s.store.Ping(ctx) }

const dateLayout = "2006-01-02"

var minPurchaseDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func runeLen(s string) int { return utf8.RuneCountInString(s) }

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}

func parseDate(v *domain.Validator, field, s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		v.Add(field, "Vui lòng nhập ngày")
		return time.Time{}
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		v.Add(field, "Ngày không hợp lệ (định dạng YYYY-MM-DD)")
		return time.Time{}
	}
	return t
}

// notFound giữ message khi lỗi đã có, ngược lại tạo lỗi 404 với msg.
func notFound(err error, msg string) error {
	if err == domain.ErrNotFound {
		return domain.NotFound(msg)
	}
	return err
}

// serviceTx là Store chạy trong transaction.
type serviceTx = repository.Store
