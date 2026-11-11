package service

import (
	"context"
	"net/mail"
	"net/url"
	"strings"

	"donghai/internal/domain"
)

// CompanyRequest là body PUT /api/admin/company.
type CompanyRequest struct {
	Name         string `json:"name"`
	Slogan       string `json:"slogan"`
	About        string `json:"about"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	WorkingHours string `json:"working_hours"`
	FacebookURL  string `json:"facebook_url"`
	TiktokURL    string `json:"tiktok_url"`
	ZaloURL      string `json:"zalo_url"`
	ShopeeURL    string `json:"shopee_url"`
}

// GetCompany trả thông tin công ty.
func (s *Service) GetCompany(ctx context.Context) (*domain.Company, error) {
	c, err := s.store.GetCompany(ctx)
	return c, notFound(err, "Chưa có thông tin công ty")
}

// ValidSocialURL: rỗng hoặc https://<host>.
func ValidSocialURL(raw string) bool {
	if raw == "" {
		return true
	}
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && len(raw) <= 500
}

// UpdateCompany ghi đè thông tin công ty.
func (s *Service) UpdateCompany(ctx context.Context, me *domain.User, in CompanyRequest) (*domain.Company, error) {
	c := &domain.Company{
		Name:         strings.TrimSpace(in.Name),
		Slogan:       strings.TrimSpace(in.Slogan),
		About:        strings.TrimSpace(in.About),
		Address:      strings.TrimSpace(in.Address),
		Phone:        strings.TrimSpace(in.Phone),
		Email:        strings.TrimSpace(in.Email),
		WorkingHours: strings.TrimSpace(in.WorkingHours),
		FacebookURL:  strings.TrimSpace(in.FacebookURL),
		TiktokURL:    strings.TrimSpace(in.TiktokURL),
		ZaloURL:      strings.TrimSpace(in.ZaloURL),
		ShopeeURL:    strings.TrimSpace(in.ShopeeURL),
		UpdatedBy:    &me.ID,
	}
	var v domain.Validator
	v.Check(c.Name != "" && runeLen(c.Name) <= 200, "name", "Tên công ty bắt buộc, tối đa 200 ký tự")
	v.Check(runeLen(c.Slogan) <= 300, "slogan", "Slogan tối đa 300 ký tự")
	v.Check(runeLen(c.About) <= 5000, "about", "Giới thiệu tối đa 5000 ký tự")
	v.Check(runeLen(c.Address) <= 300, "address", "Địa chỉ tối đa 300 ký tự")
	v.Check(runeLen(c.Phone) <= 50, "phone", "Số điện thoại tối đa 50 ký tự")
	v.Check(runeLen(c.WorkingHours) <= 200, "working_hours", "Giờ làm việc tối đa 200 ký tự")
	if c.Email != "" {
		_, err := mail.ParseAddress(c.Email)
		v.Check(err == nil && runeLen(c.Email) <= 200, "email", "Email không hợp lệ")
	}
	for field, val := range map[string]string{
		"facebook_url": c.FacebookURL, "tiktok_url": c.TiktokURL, "zalo_url": c.ZaloURL, "shopee_url": c.ShopeeURL,
	} {
		v.Check(ValidSocialURL(val), field, "Đường dẫn phải bắt đầu bằng https:// hoặc để trống")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateCompany(ctx, c); err != nil {
		return nil, err
	}
	return s.store.GetCompany(ctx)
}
