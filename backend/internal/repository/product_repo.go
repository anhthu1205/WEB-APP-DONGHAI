package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

// Không bao giờ SELECT cột data của product_images ở đây.
const productCols = `p.id, p.name, p.slug, p.sku, p.short_description, p.description, p.specs, p.unit,
	p.price_public, p.price_internal, p.is_visible, p.sort_order, p.created_at, p.updated_at,
	(SELECT i.id::text FROM product_images i WHERE i.product_id = p.id
	   ORDER BY i.is_main DESC, i.sort_order, i.created_at LIMIT 1) AS main_image_id`

func scanProduct(r pgx.Row) (*domain.Product, error) {
	var p domain.Product
	var specs []byte
	err := r.Scan(&p.ID, &p.Name, &p.Slug, &p.SKU, &p.ShortDescription, &p.Description, &specs, &p.Unit,
		&p.PricePublic, &p.PriceInternal, &p.IsVisible, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt, &p.MainImageID)
	if err != nil {
		return nil, mapErr(err)
	}
	p.Specs = []domain.Spec{}
	if len(specs) > 0 {
		if err := json.Unmarshal(specs, &p.Specs); err != nil {
			return nil, fmt.Errorf("specs: %w", err)
		}
	}
	return &p, nil
}

// ListProducts liệt kê SP theo sort_order, id; onlyVisible cho trang công khai; q lọc theo tên/SKU.
func (s *Store) ListProducts(ctx context.Context, q string, onlyVisible bool) ([]*domain.Product, error) {
	sql := `SELECT ` + productCols + ` FROM products p WHERE ($1 = false OR p.is_visible)
		AND ($2 = '' OR p.name ILIKE $3 OR p.sku ILIKE $3 OR p.slug ILIKE $3)
		ORDER BY p.sort_order, p.id`
	rows, err := s.q.Query(ctx, sql, onlyVisible, q, likePattern(q))
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(r pgx.Rows) (*domain.Product, error) { return scanProduct(r) })
}

// GetProductByID lấy SP theo id (kể cả ẩn).
func (s *Store) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return scanProduct(s.q.QueryRow(ctx, `SELECT `+productCols+` FROM products p WHERE p.id = $1`, id))
}

// GetVisibleProductBySlug lấy SP đang hiển thị theo slug.
func (s *Store) GetVisibleProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return scanProduct(s.q.QueryRow(ctx, `SELECT `+productCols+` FROM products p WHERE p.slug = $1 AND p.is_visible`, slug))
}

// GetProductsByIDs trả map id → SP (dùng khi tạo đơn/KM).
func (s *Store) GetProductsByIDs(ctx context.Context, ids []int64) (map[int64]*domain.Product, error) {
	rows, err := s.q.Query(ctx, `SELECT `+productCols+` FROM products p WHERE p.id = ANY($1)`, ids)
	if err != nil {
		return nil, mapErr(err)
	}
	list, err := collect(rows, func(r pgx.Rows) (*domain.Product, error) { return scanProduct(r) })
	if err != nil {
		return nil, err
	}
	m := make(map[int64]*domain.Product, len(list))
	for _, p := range list {
		m[p.ID] = p
	}
	return m, nil
}

// SlugExists kiểm tra slug đã dùng bởi SP khác excludeID.
func (s *Store) SlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var ok bool
	err := s.q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM products WHERE slug = $1 AND id <> $2)`, slug, excludeID).Scan(&ok)
	return ok, mapErr(err)
}

func specsJSON(sp []domain.Spec) []byte {
	if sp == nil {
		sp = []domain.Spec{}
	}
	b, _ := json.Marshal(sp)
	return b
}

// InsertProduct thêm SP, trả id.
func (s *Store) InsertProduct(ctx context.Context, p *domain.Product) (int64, error) {
	var id int64
	err := s.q.QueryRow(ctx, `INSERT INTO products (name, slug, sku, short_description, description, specs, unit,
		price_public, price_internal, is_visible, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		p.Name, p.Slug, p.SKU, p.ShortDescription, p.Description, specsJSON(p.Specs), p.Unit,
		p.PricePublic, p.PriceInternal, p.IsVisible, p.SortOrder).Scan(&id)
	return id, mapErr(err)
}

// UpdateProduct ghi đè toàn bộ trường SP.
func (s *Store) UpdateProduct(ctx context.Context, p *domain.Product) error {
	tag, err := s.q.Exec(ctx, `UPDATE products SET name = $2, slug = $3, sku = $4, short_description = $5,
		description = $6, specs = $7, unit = $8, price_public = $9, price_internal = $10, is_visible = $11,
		sort_order = $12 WHERE id = $1`,
		p.ID, p.Name, p.Slug, p.SKU, p.ShortDescription, p.Description, specsJSON(p.Specs), p.Unit,
		p.PricePublic, p.PriceInternal, p.IsVisible, p.SortOrder)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetProductVisibility ẩn/hiện SP.
func (s *Store) SetProductVisibility(ctx context.Context, id int64, visible bool) error {
	tag, err := s.q.Exec(ctx, `UPDATE products SET is_visible = $2 WHERE id = $1`, id, visible)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ProductHasOrders kiểm tra SP đã có trong đơn hàng.
func (s *Store) ProductHasOrders(ctx context.Context, id int64) (bool, error) {
	var ok bool
	err := s.q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM order_items WHERE product_id = $1)`, id).Scan(&ok)
	return ok, mapErr(err)
}

// DeleteProduct xóa SP (ảnh/KM item xóa theo cascade).
func (s *Store) DeleteProduct(ctx context.Context, id int64) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
