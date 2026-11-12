package service

import (
	"context"
	"fmt"
	"strings"

	"donghai/internal/domain"
	"donghai/internal/repository"
	"donghai/internal/slug"
)

// CategoryRequest là body POST/PUT /api/admin/categories.
type CategoryRequest struct {
	Name        string  `json:"name"`
	Slug        *string `json:"slug"` // bỏ trống → sinh từ tên (tạo mới) hoặc giữ slug cũ (sửa)
	Description string  `json:"description"`
	IsVisible   *bool   `json:"is_visible"`   // mặc định true
	ShowInMenu  *bool   `json:"show_in_menu"` // mặc định true
	ShowOnHome  *bool   `json:"show_on_home"` // mặc định false
	SortOrder   int     `json:"sort_order"`
}

// CategoryPage là danh mục kèm SP để hiển thị trang danh mục.
type CategoryPage struct {
	Category *domain.Category
	Products []*domain.Product
}

func boolOr(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func buildCategory(in CategoryRequest) (*domain.Category, bool, error) {
	c := &domain.Category{
		Name:        strings.Join(strings.Fields(in.Name), " "),
		Description: strings.TrimSpace(in.Description),
		IsVisible:   boolOr(in.IsVisible, true),
		ShowInMenu:  boolOr(in.ShowInMenu, true),
		ShowOnHome:  boolOr(in.ShowOnHome, false),
		SortOrder:   in.SortOrder,
	}
	var v domain.Validator
	v.Check(c.Name != "" && runeLen(c.Name) <= 100, "name", "Tên danh mục bắt buộc, tối đa 100 ký tự")
	v.Check(runeLen(c.Description) <= 2000, "description", "Mô tả tối đa 2000 ký tự")
	v.Check(c.SortOrder >= -100000 && c.SortOrder <= 100000, "sort_order", "Thứ tự không hợp lệ")
	explicit := false
	if sl := trimPtr(in.Slug); sl != nil {
		explicit = true
		c.Slug = strings.ToLower(*sl)
		v.Check(slug.Valid.MatchString(c.Slug) && len(c.Slug) <= 120, "slug", "Đường dẫn chỉ gồm chữ thường không dấu, số và dấu gạch nối")
	}
	return c, explicit, v.Err()
}

// uniqueCategorySlug sinh slug từ tên, thêm -2, -3… nếu trùng.
func (s *Service) uniqueCategorySlug(ctx context.Context, name string, excludeID int64) (string, error) {
	base := slug.Make(name)
	if base == "" {
		base = "danh-muc"
	}
	for i := 1; i <= 1000; i++ {
		cand := base
		if i > 1 {
			cand = fmt.Sprintf("%s-%d", base, i)
		}
		exists, err := s.store.CategorySlugExists(ctx, cand, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return cand, nil
		}
	}
	return "", domain.Conflict("Không sinh được đường dẫn duy nhất")
}

// ListPublicCategories: danh mục đang hiển thị (product_count = SP khách thấy được).
func (s *Service) ListPublicCategories(ctx context.Context) ([]*domain.Category, error) {
	return s.store.ListCategories(ctx, true)
}

// ListInternalCategories: tất cả danh mục (product_count = tất cả SP).
func (s *Service) ListInternalCategories(ctx context.Context) ([]*domain.Category, error) {
	return s.store.ListCategories(ctx, false)
}

// GetCategoryPage: danh mục đang hiển thị theo slug + SP khách thấy được (danh mục ẩn → 404).
func (s *Service) GetCategoryPage(ctx context.Context, slugStr string) (*CategoryPage, error) {
	c, err := s.store.GetVisibleCategoryBySlug(ctx, slugStr)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy danh mục")
	}
	list, err := s.store.ListProducts(ctx, repository.ProductFilter{OnlyVisible: true, Scope: repository.ScopeCategory, CategoryID: c.ID})
	if err != nil {
		return nil, err
	}
	return &CategoryPage{Category: c, Products: list}, nil
}

// CreateCategory thêm danh mục (admin).
func (s *Service) CreateCategory(ctx context.Context, in CategoryRequest) (*domain.Category, error) {
	c, explicit, err := buildCategory(in)
	if err != nil {
		return nil, err
	}
	if !explicit {
		if c.Slug, err = s.uniqueCategorySlug(ctx, c.Name, 0); err != nil {
			return nil, err
		}
	}
	id, err := s.store.InsertCategory(ctx, c)
	if err != nil {
		return nil, err
	}
	return s.store.GetCategory(ctx, id)
}

// UpdateCategory sửa danh mục (admin). Không gửi slug → giữ slug cũ.
func (s *Service) UpdateCategory(ctx context.Context, id int64, in CategoryRequest) (*domain.Category, error) {
	old, err := s.store.GetCategory(ctx, id)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy danh mục")
	}
	c, explicit, err := buildCategory(in)
	if err != nil {
		return nil, err
	}
	c.ID = id
	if !explicit {
		c.Slug = old.Slug
	}
	if err := s.store.UpdateCategory(ctx, c); err != nil {
		return nil, notFound(err, "Không tìm thấy danh mục")
	}
	return s.store.GetCategory(ctx, id)
}

// DeleteCategory xóa danh mục không còn SP (admin).
func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return notFound(s.store.DeleteCategory(ctx, id), "Không tìm thấy danh mục")
}

// UploadCategoryCover thay ảnh bìa danh mục (admin). Chỉ JPEG/PNG/WebP thật, ≤ 2 MB.
func (s *Service) UploadCategoryCover(ctx context.Context, id int64, data []byte) (*domain.Category, error) {
	mime, _, _, err := DetectImage(data)
	if err != nil {
		return nil, err
	}
	ver := LogoVersion(data)
	if err := s.store.SetCategoryCover(ctx, id, &mime, &ver, data); err != nil {
		return nil, notFound(err, "Không tìm thấy danh mục")
	}
	return s.store.GetCategory(ctx, id)
}

// DeleteCategoryCover xóa ảnh bìa danh mục (admin).
func (s *Service) DeleteCategoryCover(ctx context.Context, id int64) (*domain.Category, error) {
	if err := s.store.SetCategoryCover(ctx, id, nil, nil, nil); err != nil {
		return nil, notFound(err, "Không tìm thấy danh mục")
	}
	return s.store.GetCategory(ctx, id)
}

// GetCategoryCover trả mime, version, bytes ảnh bìa.
func (s *Service) GetCategoryCover(ctx context.Context, id int64) (string, string, []byte, error) {
	mime, ver, data, err := s.store.GetCategoryCover(ctx, id)
	return mime, ver, data, notFound(err, "Chưa có ảnh bìa")
}
