package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"donghai/internal/domain"
	"donghai/internal/slug"
)

// MaxPrice là giá tối đa chấp nhận (1.000 tỷ VND) để tránh nhập nhầm.
const MaxPrice int64 = 1_000_000_000_000

// ProductRequest là body POST/PUT /api/admin/products.
type ProductRequest struct {
	Name             string        `json:"name"`
	Slug             *string       `json:"slug"`
	SKU              *string       `json:"sku"`
	ShortDescription string        `json:"short_description"`
	Description      string        `json:"description"`
	Specs            []domain.Spec `json:"specs"`
	Unit             string        `json:"unit"`
	PricePublic      int64         `json:"price_public"`
	PriceInternal    *int64        `json:"price_internal"`
	IsVisible        *bool         `json:"is_visible"`
	SortOrder        int           `json:"sort_order"`
}

// VisibilityRequest là body PATCH /api/admin/products/{id}/visibility.
type VisibilityRequest struct {
	IsVisible *bool `json:"is_visible"`
}

// ListPublicProducts: SP đang hiển thị.
func (s *Service) ListPublicProducts(ctx context.Context, q string) ([]*domain.Product, error) {
	return s.store.ListProducts(ctx, strings.TrimSpace(q), true)
}

// GetPublicProduct: chi tiết SP đang hiển thị theo slug (ẩn → 404).
func (s *Service) GetPublicProduct(ctx context.Context, slugStr string) (*domain.Product, error) {
	p, err := s.store.GetVisibleProductBySlug(ctx, slugStr)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy sản phẩm")
	}
	if p.Images, err = s.store.ListImages(ctx, p.ID); err != nil {
		return nil, err
	}
	return p, nil
}

// ListInternalProducts: tất cả SP (kể cả ẩn).
func (s *Service) ListInternalProducts(ctx context.Context, q string) ([]*domain.Product, error) {
	return s.store.ListProducts(ctx, strings.TrimSpace(q), false)
}

// GetInternalProduct: chi tiết SP kèm ảnh.
func (s *Service) GetInternalProduct(ctx context.Context, id int64) (*domain.Product, error) {
	p, err := s.store.GetProductByID(ctx, id)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy sản phẩm")
	}
	if p.Images, err = s.store.ListImages(ctx, p.ID); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) buildProduct(in ProductRequest) (*domain.Product, bool, error) {
	p := &domain.Product{
		Name:             strings.TrimSpace(in.Name),
		SKU:              trimPtr(in.SKU),
		ShortDescription: strings.TrimSpace(in.ShortDescription),
		Description:      strings.TrimSpace(in.Description),
		Unit:             strings.TrimSpace(in.Unit),
		PricePublic:      in.PricePublic,
		PriceInternal:    in.PriceInternal,
		IsVisible:        in.IsVisible == nil || *in.IsVisible,
		SortOrder:        in.SortOrder,
	}
	if p.Unit == "" {
		p.Unit = "cái"
	}
	var v domain.Validator
	v.Check(p.Name != "" && runeLen(p.Name) <= 200, "name", "Tên sản phẩm bắt buộc, tối đa 200 ký tự")
	v.Check(runeLen(p.ShortDescription) <= 500, "short_description", "Mô tả ngắn tối đa 500 ký tự")
	v.Check(runeLen(p.Description) <= 20000, "description", "Mô tả chi tiết tối đa 20000 ký tự")
	v.Check(runeLen(p.Unit) <= 20, "unit", "Đơn vị tối đa 20 ký tự")
	v.Check(p.PricePublic >= 0 && p.PricePublic <= MaxPrice, "price_public", "Giá niêm yết không hợp lệ")
	v.Check(p.PriceInternal == nil || (*p.PriceInternal >= 0 && *p.PriceInternal <= MaxPrice), "price_internal", "Giá nội bộ không hợp lệ")
	v.Check(p.SortOrder >= -100000 && p.SortOrder <= 100000, "sort_order", "Thứ tự không hợp lệ")
	if p.SKU != nil {
		v.Check(runeLen(*p.SKU) <= 50, "sku", "SKU tối đa 50 ký tự")
	}
	v.Check(len(in.Specs) <= 50, "specs", "Tối đa 50 thông số")
	p.Specs = make([]domain.Spec, 0, len(in.Specs))
	for i, sp := range in.Specs {
		l, val := strings.TrimSpace(sp.Label), strings.TrimSpace(sp.Value)
		if l == "" && val == "" {
			continue
		}
		v.Check(l != "" && runeLen(l) <= 100, fmt.Sprintf("specs.%d.label", i), "Tên thông số bắt buộc, tối đa 100 ký tự")
		v.Check(runeLen(val) <= 500, fmt.Sprintf("specs.%d.value", i), "Giá trị thông số tối đa 500 ký tự")
		p.Specs = append(p.Specs, domain.Spec{Label: l, Value: val})
	}
	explicitSlug := false
	if sl := trimPtr(in.Slug); sl != nil {
		explicitSlug = true
		p.Slug = strings.ToLower(*sl)
		v.Check(slug.Valid.MatchString(p.Slug) && len(p.Slug) <= 120, "slug", "Slug chỉ gồm chữ thường không dấu, số và dấu gạch nối")
	}
	return p, explicitSlug, v.Err()
}

// uniqueSlug sinh slug từ tên, thêm -2, -3… nếu trùng.
func (s *Service) uniqueSlug(ctx context.Context, name string, excludeID int64) (string, error) {
	base := slug.Make(name)
	if base == "" {
		base = "san-pham"
	}
	for i := 1; i <= 1000; i++ {
		cand := base
		if i > 1 {
			cand = fmt.Sprintf("%s-%d", base, i)
		}
		exists, err := s.store.SlugExists(ctx, cand, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return cand, nil
		}
	}
	return "", domain.Conflict("Không sinh được slug duy nhất")
}

// CreateProduct thêm SP.
func (s *Service) CreateProduct(ctx context.Context, in ProductRequest) (*domain.Product, error) {
	p, explicit, err := s.buildProduct(in)
	if err != nil {
		return nil, err
	}
	if !explicit {
		if p.Slug, err = s.uniqueSlug(ctx, p.Name, 0); err != nil {
			return nil, err
		}
	}
	id, err := s.store.InsertProduct(ctx, p)
	if err != nil {
		return nil, err
	}
	return s.GetInternalProduct(ctx, id)
}

// UpdateProduct cập nhật toàn bộ SP. Không gửi slug → giữ slug cũ.
func (s *Service) UpdateProduct(ctx context.Context, id int64, in ProductRequest) (*domain.Product, error) {
	old, err := s.store.GetProductByID(ctx, id)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy sản phẩm")
	}
	p, explicit, err := s.buildProduct(in)
	if err != nil {
		return nil, err
	}
	p.ID = id
	if !explicit {
		p.Slug = old.Slug
	}
	if err := s.store.UpdateProduct(ctx, p); err != nil {
		return nil, notFound(err, "Không tìm thấy sản phẩm")
	}
	return s.GetInternalProduct(ctx, id)
}

// SetProductVisibility ẩn/hiện SP.
func (s *Service) SetProductVisibility(ctx context.Context, id int64, in VisibilityRequest) error {
	if in.IsVisible == nil {
		return &domain.ValidationError{Fields: map[string]string{"is_visible": "Thiếu trường is_visible"}}
	}
	return notFound(s.store.SetProductVisibility(ctx, id, *in.IsVisible), "Không tìm thấy sản phẩm")
}

// DeleteProduct xóa SP chưa có đơn hàng.
func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	return s.store.InTx(ctx, func(tx *serviceTx) error {
		has, err := tx.ProductHasOrders(ctx, id)
		if err != nil {
			return err
		}
		if has {
			return domain.Conflict("Sản phẩm đã có đơn hàng, chỉ có thể ẩn")
		}
		err = tx.DeleteProduct(ctx, id)
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NotFound("Không tìm thấy sản phẩm")
		}
		return err
	})
}
