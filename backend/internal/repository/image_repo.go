package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

const imageMetaCols = `id::text, product_id, mime_type, size_bytes, width, height, is_main, sort_order, created_at`

func scanImageMeta(r pgx.Row) (*domain.ProductImage, error) {
	var im domain.ProductImage
	err := r.Scan(&im.ID, &im.ProductID, &im.MimeType, &im.SizeBytes, &im.Width, &im.Height,
		&im.IsMain, &im.SortOrder, &im.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &im, nil
}

// ListImages liệt kê metadata ảnh của SP (không lấy data).
func (s *Store) ListImages(ctx context.Context, productID int64) ([]domain.ProductImage, error) {
	rows, err := s.q.Query(ctx, `SELECT `+imageMetaCols+` FROM product_images WHERE product_id = $1
		ORDER BY is_main DESC, sort_order, created_at`, productID)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(r pgx.Rows) (domain.ProductImage, error) {
		im, err := scanImageMeta(r)
		if err != nil {
			return domain.ProductImage{}, err
		}
		return *im, nil
	})
}

// GetImageMeta lấy metadata 1 ảnh.
func (s *Store) GetImageMeta(ctx context.Context, id string) (*domain.ProductImage, error) {
	return scanImageMeta(s.q.QueryRow(ctx, `SELECT `+imageMetaCols+` FROM product_images WHERE id = $1`, id))
}

// GetImageData lấy mime + bytes ảnh.
func (s *Store) GetImageData(ctx context.Context, id string) (string, []byte, error) {
	var mime string
	var data []byte
	err := s.q.QueryRow(ctx, `SELECT mime_type, data FROM product_images WHERE id = $1`, id).Scan(&mime, &data)
	return mime, data, mapErr(err)
}

// LockProductForImages khóa dòng SP (tránh 2 upload song song vượt giới hạn) và trả số ảnh + sort lớn nhất.
func (s *Store) LockProductForImages(ctx context.Context, productID int64) (count, maxSort int, err error) {
	var id int64
	if err = s.q.QueryRow(ctx, `SELECT id FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&id); err != nil {
		return 0, 0, mapErr(err)
	}
	err = s.q.QueryRow(ctx, `SELECT count(*), COALESCE(max(sort_order), 0) FROM product_images WHERE product_id = $1`,
		productID).Scan(&count, &maxSort)
	return count, maxSort, mapErr(err)
}

// InsertImage thêm ảnh, trả metadata.
func (s *Store) InsertImage(ctx context.Context, im *domain.ProductImage) (*domain.ProductImage, error) {
	return scanImageMeta(s.q.QueryRow(ctx, `INSERT INTO product_images
		(product_id, data, mime_type, size_bytes, width, height, is_main, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+imageMetaCols,
		im.ProductID, im.Data, im.MimeType, im.SizeBytes, im.Width, im.Height, im.IsMain, im.SortOrder))
}

// SetMainImage đặt ảnh id làm ảnh chính của SP (bỏ chính ảnh cũ trước).
func (s *Store) SetMainImage(ctx context.Context, productID int64, id string) error {
	if _, err := s.q.Exec(ctx, `UPDATE product_images SET is_main = false WHERE product_id = $1 AND is_main AND id <> $2`,
		productID, id); err != nil {
		return mapErr(err)
	}
	_, err := s.q.Exec(ctx, `UPDATE product_images SET is_main = true WHERE id = $1`, id)
	return mapErr(err)
}

// DeleteImage xóa ảnh.
func (s *Store) DeleteImage(ctx context.Context, id string) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM product_images WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// PromoteFirstImage: nếu SP không còn ảnh chính → ảnh sort_order nhỏ nhất thành chính.
func (s *Store) PromoteFirstImage(ctx context.Context, productID int64) error {
	_, err := s.q.Exec(ctx, `UPDATE product_images SET is_main = true WHERE id = (
		SELECT id FROM product_images WHERE product_id = $1 ORDER BY sort_order, created_at LIMIT 1)
		AND NOT EXISTS (SELECT 1 FROM product_images WHERE product_id = $1 AND is_main)`, productID)
	return mapErr(err)
}

// CopyImages sao chép toàn bộ ảnh (cả dữ liệu, ảnh chính, thứ tự) từ SP from sang SP to; trả số ảnh đã chép.
func (s *Store) CopyImages(ctx context.Context, from, to int64) (int, error) {
	tag, err := s.q.Exec(ctx, `INSERT INTO product_images (product_id, data, mime_type, size_bytes, width, height, is_main, sort_order)
		SELECT $2, data, mime_type, size_bytes, width, height, is_main, sort_order FROM product_images WHERE product_id = $1`,
		from, to)
	if err != nil {
		return 0, mapErr(err)
	}
	return int(tag.RowsAffected()), nil
}
