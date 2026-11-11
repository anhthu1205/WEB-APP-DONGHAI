package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"donghai/internal/domain"
	"donghai/internal/repository"
)

// PromotionItemRequest là 1 SP trong KM.
type PromotionItemRequest struct {
	ProductID  int64  `json:"product_id"`
	PromoPrice int64  `json:"promo_price"`
	Note       string `json:"note"`
}

// PromotionRequest là body POST/PUT /api/admin/promotions.
type PromotionRequest struct {
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	StartDate string                 `json:"start_date"`
	EndDate   string                 `json:"end_date"`
	IsActive  *bool                  `json:"is_active"`
	Items     []PromotionItemRequest `json:"items"`
}

// PriceList là dữ liệu trang ẩn bảng giá nội bộ.
type PriceList struct {
	Today      time.Time
	Promotions []*domain.Promotion
	Products   []domain.PriceListProduct
}

// PromotionStatus: active | upcoming | ended theo ngày (không xét is_active).
func PromotionStatus(p *domain.Promotion, today time.Time) string {
	switch {
	case today.Before(p.StartDate):
		return "upcoming"
	case today.After(p.EndDate):
		return "ended"
	default:
		return "active"
	}
}

// GetPriceList: KM đang chạy + sắp tới (is_active) và bảng giá kèm giá KM hiện hành.
func (s *Service) GetPriceList(ctx context.Context) (*PriceList, error) {
	today := s.Today()
	promos, err := s.store.ListCurrentAndUpcomingPromotions(ctx, today)
	if err != nil {
		return nil, err
	}
	products, err := s.store.PriceListProducts(ctx, today)
	if err != nil {
		return nil, err
	}
	return &PriceList{Today: today, Promotions: promos, Products: products}, nil
}

// GetPromotion: chi tiết KM.
func (s *Service) GetPromotion(ctx context.Context, id int64) (*domain.Promotion, error) {
	p, err := s.store.GetPromotion(ctx, id)
	return p, notFound(err, "Không tìm thấy chương trình khuyến mãi")
}

// ListPromotions: danh sách KM (admin), lọc status.
func (s *Service) ListPromotions(ctx context.Context, status string, pp domain.PageParams) (*domain.Page[*domain.Promotion], error) {
	switch status {
	case "", "active", "upcoming", "ended":
	default:
		return nil, &domain.ValidationError{Fields: map[string]string{"status": "Trạng thái phải là active, upcoming hoặc ended"}}
	}
	return s.store.ListPromotions(ctx, repository.PromotionFilter{Status: status, Today: s.Today()}, pp)
}

func (s *Service) buildPromotion(ctx context.Context, in PromotionRequest) (*domain.Promotion, error) {
	p := &domain.Promotion{
		Title:    strings.TrimSpace(in.Title),
		Content:  strings.TrimSpace(in.Content),
		IsActive: in.IsActive == nil || *in.IsActive,
	}
	var v domain.Validator
	v.Check(p.Title != "" && runeLen(p.Title) <= 200, "title", "Tiêu đề bắt buộc, tối đa 200 ký tự")
	v.Check(runeLen(p.Content) <= 10000, "content", "Nội dung tối đa 10000 ký tự")
	p.StartDate = parseDate(&v, "start_date", in.StartDate)
	p.EndDate = parseDate(&v, "end_date", in.EndDate)
	if !p.StartDate.IsZero() && !p.EndDate.IsZero() && p.EndDate.Before(p.StartDate) {
		v.Add("end_date", "Ngày kết thúc phải sau hoặc bằng ngày bắt đầu")
	}
	v.Check(len(in.Items) <= 200, "items", "Tối đa 200 sản phẩm")
	seen := map[int64]bool{}
	ids := []int64{}
	for i, it := range in.Items {
		if seen[it.ProductID] {
			v.Add(fmt.Sprintf("items.%d.product_id", i), "Sản phẩm bị lặp")
			continue
		}
		seen[it.ProductID] = true
		ids = append(ids, it.ProductID)
		v.Check(it.PromoPrice >= 0 && it.PromoPrice <= MaxPrice, fmt.Sprintf("items.%d.promo_price", i), "Giá khuyến mãi không hợp lệ")
		note := strings.TrimSpace(it.Note)
		v.Check(runeLen(note) <= 200, fmt.Sprintf("items.%d.note", i), "Ghi chú tối đa 200 ký tự")
		p.Items = append(p.Items, domain.PromotionItem{ProductID: it.ProductID, PromoPrice: it.PromoPrice, Note: note})
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		m, err := s.store.GetProductsByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		for i, it := range p.Items {
			if m[it.ProductID] == nil {
				v.Add(fmt.Sprintf("items.%d.product_id", i), "Sản phẩm không tồn tại")
			}
		}
	}
	return p, v.Err()
}

// CreatePromotion thêm KM kèm items (1 transaction).
func (s *Service) CreatePromotion(ctx context.Context, me *domain.User, in PromotionRequest) (*domain.Promotion, error) {
	p, err := s.buildPromotion(ctx, in)
	if err != nil {
		return nil, err
	}
	p.CreatedBy = &me.ID
	var id int64
	err = s.store.InTx(ctx, func(tx *serviceTx) error {
		if id, err = tx.InsertPromotion(ctx, p); err != nil {
			return err
		}
		return tx.ReplacePromotionItems(ctx, id, p.Items)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPromotion(ctx, id)
}

// UpdatePromotion thay toàn bộ KM và items (1 transaction).
func (s *Service) UpdatePromotion(ctx context.Context, id int64, in PromotionRequest) (*domain.Promotion, error) {
	p, err := s.buildPromotion(ctx, in)
	if err != nil {
		return nil, err
	}
	p.ID = id
	err = s.store.InTx(ctx, func(tx *serviceTx) error {
		if err := tx.UpdatePromotion(ctx, p); err != nil {
			return notFound(err, "Không tìm thấy chương trình khuyến mãi")
		}
		return tx.ReplacePromotionItems(ctx, id, p.Items)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPromotion(ctx, id)
}

// DeletePromotion xóa KM.
func (s *Service) DeletePromotion(ctx context.Context, id int64) error {
	return notFound(s.store.DeletePromotion(ctx, id), "Không tìm thấy chương trình khuyến mãi")
}
