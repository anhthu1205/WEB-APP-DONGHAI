package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

const promoCols = `pr.id, pr.title, pr.content, pr.start_date, pr.end_date, pr.is_active, pr.created_by,
	pr.created_at, pr.updated_at, (SELECT count(*) FROM promotion_items pi WHERE pi.promotion_id = pr.id)::int`

func scanPromotion(r pgx.Row) (*domain.Promotion, error) {
	var p domain.Promotion
	err := r.Scan(&p.ID, &p.Title, &p.Content, &p.StartDate, &p.EndDate, &p.IsActive, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt, &p.ItemsCount)
	if err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

// PromotionFilter lọc danh sách KM theo trạng thái thời gian.
type PromotionFilter struct {
	Status string // "", active, upcoming, ended
	Today  time.Time
}

// ListPromotions liệt kê KM có phân trang.
func (s *Store) ListPromotions(ctx context.Context, f PromotionFilter, pp domain.PageParams) (*domain.Page[*domain.Promotion], error) {
	where := `WHERE CASE $1::text
		WHEN 'active' THEN pr.start_date <= $2::date AND pr.end_date >= $2::date
		WHEN 'upcoming' THEN pr.start_date > $2::date
		WHEN 'ended' THEN pr.end_date < $2::date
		ELSE true END`
	var total int
	if err := s.q.QueryRow(ctx, `SELECT count(*) FROM promotions pr `+where, f.Status, f.Today).Scan(&total); err != nil {
		return nil, mapErr(err)
	}
	rows, err := s.q.Query(ctx, `SELECT `+promoCols+` FROM promotions pr `+where+`
		ORDER BY pr.start_date DESC, pr.id DESC LIMIT $3 OFFSET $4`, f.Status, f.Today, pp.PageSize, pp.Offset())
	if err != nil {
		return nil, mapErr(err)
	}
	items, err := collect(rows, func(r pgx.Rows) (*domain.Promotion, error) { return scanPromotion(r) })
	if err != nil {
		return nil, err
	}
	return &domain.Page[*domain.Promotion]{Items: items, Total: total, Page: pp.Page, PageSize: pp.PageSize}, nil
}

// GetPromotion lấy KM kèm items.
func (s *Store) GetPromotion(ctx context.Context, id int64) (*domain.Promotion, error) {
	p, err := scanPromotion(s.q.QueryRow(ctx, `SELECT `+promoCols+` FROM promotions pr WHERE pr.id = $1`, id))
	if err != nil {
		return nil, err
	}
	m, err := s.promotionItems(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	p.Items = m[id]
	return p, nil
}

// ListCurrentAndUpcomingPromotions: KM is_active, end_date >= today, kèm items.
func (s *Store) ListCurrentAndUpcomingPromotions(ctx context.Context, today time.Time) ([]*domain.Promotion, error) {
	rows, err := s.q.Query(ctx, `SELECT `+promoCols+` FROM promotions pr
		WHERE pr.is_active AND pr.end_date >= $1::date ORDER BY pr.start_date, pr.id`, today)
	if err != nil {
		return nil, mapErr(err)
	}
	list, err := collect(rows, func(r pgx.Rows) (*domain.Promotion, error) { return scanPromotion(r) })
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(list))
	for i, p := range list {
		ids[i] = p.ID
	}
	m, err := s.promotionItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		p.Items = m[p.ID]
	}
	return list, nil
}

func (s *Store) promotionItems(ctx context.Context, ids []int64) (map[int64][]domain.PromotionItem, error) {
	out := map[int64][]domain.PromotionItem{}
	for _, id := range ids {
		out[id] = []domain.PromotionItem{}
	}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.q.Query(ctx, `SELECT pi.promotion_id, pi.product_id, p.name, p.price_public, pi.promo_price, pi.note
		FROM promotion_items pi JOIN products p ON p.id = pi.product_id
		WHERE pi.promotion_id = ANY($1) ORDER BY p.sort_order, p.id`, ids)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var it domain.PromotionItem
		if err := rows.Scan(&pid, &it.ProductID, &it.ProductName, &it.PricePublic, &it.PromoPrice, &it.Note); err != nil {
			return nil, err
		}
		out[pid] = append(out[pid], it)
	}
	return out, rows.Err()
}

// InsertPromotion thêm KM (không gồm items), trả id.
func (s *Store) InsertPromotion(ctx context.Context, p *domain.Promotion) (int64, error) {
	var id int64
	err := s.q.QueryRow(ctx, `INSERT INTO promotions (title, content, start_date, end_date, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		p.Title, p.Content, p.StartDate, p.EndDate, p.IsActive, p.CreatedBy).Scan(&id)
	return id, mapErr(err)
}

// UpdatePromotion ghi đè KM.
func (s *Store) UpdatePromotion(ctx context.Context, p *domain.Promotion) error {
	tag, err := s.q.Exec(ctx, `UPDATE promotions SET title = $2, content = $3, start_date = $4, end_date = $5,
		is_active = $6 WHERE id = $1`, p.ID, p.Title, p.Content, p.StartDate, p.EndDate, p.IsActive)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ReplacePromotionItems thay toàn bộ items của KM.
func (s *Store) ReplacePromotionItems(ctx context.Context, promotionID int64, items []domain.PromotionItem) error {
	if _, err := s.q.Exec(ctx, `DELETE FROM promotion_items WHERE promotion_id = $1`, promotionID); err != nil {
		return mapErr(err)
	}
	for _, it := range items {
		if _, err := s.q.Exec(ctx, `INSERT INTO promotion_items (promotion_id, product_id, promo_price, note)
			VALUES ($1, $2, $3, $4)`, promotionID, it.ProductID, it.PromoPrice, it.Note); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

// DeletePromotion xóa KM.
func (s *Store) DeletePromotion(ctx context.Context, id int64) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM promotions WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// PriceListProducts: SP đang hiển thị kèm giá KM thấp nhất đang hiệu lực hôm nay.
func (s *Store) PriceListProducts(ctx context.Context, today time.Time) ([]domain.PriceListProduct, error) {
	rows, err := s.q.Query(ctx, `SELECT p.id, p.name, p.unit, p.price_public, p.price_internal, cur.promo_price, cur.title
		FROM products p
		LEFT JOIN LATERAL (
			SELECT pi.promo_price, pr.title FROM promotion_items pi
			JOIN promotions pr ON pr.id = pi.promotion_id
			WHERE pi.product_id = p.id AND pr.is_active AND pr.start_date <= $1::date AND pr.end_date >= $1::date
			ORDER BY pi.promo_price, pr.id LIMIT 1
		) cur ON true
		WHERE p.is_visible
		ORDER BY p.sort_order, p.id`, today)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(r pgx.Rows) (domain.PriceListProduct, error) {
		var p domain.PriceListProduct
		err := r.Scan(&p.ID, &p.Name, &p.Unit, &p.PricePublic, &p.PriceInternal, &p.CurrentPromoPrice, &p.CurrentPromotionTitle)
		return p, err
	})
}
