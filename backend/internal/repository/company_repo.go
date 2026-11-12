package repository

import (
	"context"

	"donghai/internal/domain"
)

const companyCols = `name, slogan, about, address, phone, email, working_hours,
	facebook_url, tiktok_url, zalo_url, shopee_url, updated_by, updated_at,
	(SELECT version FROM site_logo WHERE id = 1)`

// GetCompany lấy thông tin công ty.
func (s *Store) GetCompany(ctx context.Context) (*domain.Company, error) {
	var c domain.Company
	err := s.q.QueryRow(ctx, `SELECT `+companyCols+` FROM company_info WHERE id = 1`).Scan(
		&c.Name, &c.Slogan, &c.About, &c.Address, &c.Phone, &c.Email, &c.WorkingHours,
		&c.FacebookURL, &c.TiktokURL, &c.ZaloURL, &c.ShopeeURL, &c.UpdatedBy, &c.UpdatedAt, &c.LogoVersion)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

// UpdateCompany ghi đè thông tin công ty (upsert dòng id=1).
func (s *Store) UpdateCompany(ctx context.Context, c *domain.Company) error {
	_, err := s.q.Exec(ctx, `INSERT INTO company_info (id, name, slogan, about, address, phone, email, working_hours,
		facebook_url, tiktok_url, zalo_url, shopee_url, updated_by)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, slogan = EXCLUDED.slogan, about = EXCLUDED.about,
		address = EXCLUDED.address, phone = EXCLUDED.phone, email = EXCLUDED.email,
		working_hours = EXCLUDED.working_hours, facebook_url = EXCLUDED.facebook_url,
		tiktok_url = EXCLUDED.tiktok_url, zalo_url = EXCLUDED.zalo_url, shopee_url = EXCLUDED.shopee_url,
		updated_by = EXCLUDED.updated_by`,
		c.Name, c.Slogan, c.About, c.Address, c.Phone, c.Email, c.WorkingHours,
		c.FacebookURL, c.TiktokURL, c.ZaloURL, c.ShopeeURL, c.UpdatedBy)
	return mapErr(err)
}

// GetLogo trả mime, version và bytes logo (chưa có → ErrNotFound).
func (s *Store) GetLogo(ctx context.Context) (mime, version string, data []byte, err error) {
	err = s.q.QueryRow(ctx, `SELECT mime_type, version, data FROM site_logo WHERE id = 1`).Scan(&mime, &version, &data)
	return mime, version, data, mapErr(err)
}

// UpsertLogo ghi logo mới (thay logo cũ).
func (s *Store) UpsertLogo(ctx context.Context, mime, version string, data []byte, by int64) error {
	_, err := s.q.Exec(ctx, `INSERT INTO site_logo (id, data, mime_type, size_bytes, version, updated_by, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, now())
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, mime_type = EXCLUDED.mime_type,
		size_bytes = EXCLUDED.size_bytes, version = EXCLUDED.version, updated_by = EXCLUDED.updated_by, updated_at = now()`,
		data, mime, len(data), version, by)
	return mapErr(err)
}

// DeleteLogo xóa logo (quay về logo mặc định). Không có logo thì không lỗi.
func (s *Store) DeleteLogo(ctx context.Context) error {
	_, err := s.q.Exec(ctx, `DELETE FROM site_logo WHERE id = 1`)
	return mapErr(err)
}
