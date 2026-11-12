package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

// categoryCols không bao giờ lấy cover_data. $1 = onlyVisible: true → product_count chỉ đếm SP khách thấy được.
const categoryCols = `c.id, c.name, c.slug, c.description, c.is_visible, c.show_in_menu, c.show_on_home, c.sort_order,
	c.cover_version, c.created_at, c.updated_at,
	(SELECT count(*) FROM products p WHERE p.category_id = c.id AND ($1 = false OR p.is_visible))::int`

func scanCategory(r pgx.Row) (*domain.Category, error) {
	var c domain.Category
	err := r.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.IsVisible, &c.ShowInMenu, &c.ShowOnHome, &c.SortOrder,
		&c.CoverVersion, &c.CreatedAt, &c.UpdatedAt, &c.ProductCount)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

// ListCategories liệt kê danh mục theo sort_order, id; onlyVisible: chỉ danh mục đang hiển thị.
func (s *Store) ListCategories(ctx context.Context, onlyVisible bool) ([]*domain.Category, error) {
	rows, err := s.q.Query(ctx, `SELECT `+categoryCols+` FROM categories c WHERE ($1 = false OR c.is_visible)
		ORDER BY c.sort_order, c.id`, onlyVisible)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(r pgx.Rows) (*domain.Category, error) { return scanCategory(r) })
}

// GetCategory lấy danh mục theo id (kể cả ẩn), đếm tất cả SP.
func (s *Store) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	return scanCategory(s.q.QueryRow(ctx, `SELECT `+categoryCols+` FROM categories c WHERE c.id = $2`, false, id))
}

// GetVisibleCategoryBySlug lấy danh mục đang hiển thị theo slug, đếm SP khách thấy được.
func (s *Store) GetVisibleCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return scanCategory(s.q.QueryRow(ctx, `SELECT `+categoryCols+` FROM categories c WHERE c.slug = $2 AND c.is_visible`, true, slug))
}

// CategorySlugExists kiểm tra slug đã dùng bởi danh mục khác excludeID.
func (s *Store) CategorySlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var ok bool
	err := s.q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE slug = $1 AND id <> $2)`, slug, excludeID).Scan(&ok)
	return ok, mapErr(err)
}

// InsertCategory thêm danh mục, trả id.
func (s *Store) InsertCategory(ctx context.Context, c *domain.Category) (int64, error) {
	var id int64
	err := s.q.QueryRow(ctx, `INSERT INTO categories (name, slug, description, is_visible, show_in_menu, show_on_home, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		c.Name, c.Slug, c.Description, c.IsVisible, c.ShowInMenu, c.ShowOnHome, c.SortOrder).Scan(&id)
	return id, mapErr(err)
}

// UpdateCategory ghi đè các trường của danh mục (không gồm ảnh bìa).
func (s *Store) UpdateCategory(ctx context.Context, c *domain.Category) error {
	tag, err := s.q.Exec(ctx, `UPDATE categories SET name = $2, slug = $3, description = $4, is_visible = $5,
		show_in_menu = $6, show_on_home = $7, sort_order = $8 WHERE id = $1`,
		c.ID, c.Name, c.Slug, c.Description, c.IsVisible, c.ShowInMenu, c.ShowOnHome, c.SortOrder)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteCategory xóa danh mục (còn SP → lỗi FK → 409).
func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CategoryExists kiểm tra danh mục id có tồn tại.
func (s *Store) CategoryExists(ctx context.Context, id int64) (bool, error) {
	var ok bool
	err := s.q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1)`, id).Scan(&ok)
	return ok, mapErr(err)
}

// SetCategoryCover ghi ảnh bìa; nil data → xóa ảnh bìa.
func (s *Store) SetCategoryCover(ctx context.Context, id int64, mime, version *string, data []byte) error {
	tag, err := s.q.Exec(ctx, `UPDATE categories SET cover_data = $2, cover_mime = $3, cover_version = $4 WHERE id = $1`,
		id, data, mime, version)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetCategoryCover trả mime, version, bytes ảnh bìa (ảnh không nhạy cảm; admin cần xem cả khi danh mục ẩn) (chưa có → ErrNotFound).
func (s *Store) GetCategoryCover(ctx context.Context, id int64) (mime, version string, data []byte, err error) {
	err = s.q.QueryRow(ctx, `SELECT cover_mime, cover_version, cover_data FROM categories
		WHERE id = $1 AND cover_data IS NOT NULL`, id).Scan(&mime, &version, &data)
	return mime, version, data, mapErr(err)
}
