// Package repository chỉ chứa truy vấn SQL tham số hóa. Không chứa logic quyền.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"donghai/internal/domain"
)

// Querier là phần chung của pgxpool.Pool và pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store truy cập dữ liệu; có thể chạy trên pool hoặc trong transaction.
type Store struct {
	q    Querier
	pool *pgxpool.Pool // nil khi đang trong transaction
}

// New tạo Store trên pool.
func New(pool *pgxpool.Pool) *Store { return &Store{q: pool, pool: pool} }

// InTx chạy fn trong 1 transaction (lồng nhau → dùng transaction hiện tại).
func (s *Store) InTx(ctx context.Context, fn func(tx *Store) error) error {
	if s.pool == nil {
		return fn(s)
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return fn(&Store{q: tx})
	})
}

// Ping kiểm tra kết nối DB.
func (s *Store) Ping(ctx context.Context) error {
	var one int
	return s.q.QueryRow(ctx, "SELECT 1").Scan(&one)
}

var conflictMessages = map[string]string{
	"users_username_key":              "Tên đăng nhập đã tồn tại",
	"products_slug_key":               "Slug sản phẩm đã tồn tại",
	"products_sku_key":                "Mã SKU đã tồn tại",
	"categories_slug_key":             "Đường dẫn danh mục đã tồn tại",
	"products_category_id_fkey":       "Danh mục còn sản phẩm, hãy chuyển sản phẩm sang danh mục khác trước khi xóa",
	"customers_phone_key":             "Số điện thoại đã thuộc về khách hàng khác",
	"uq_product_images_main":          "Sản phẩm đã có ảnh chính",
	"order_items_product_id_fkey":     "Sản phẩm đã có đơn hàng, chỉ có thể ẩn",
	"orders_customer_id_fkey":         "Khách hàng đã có đơn hàng, không thể xóa",
	"promotion_items_pkey":            "Sản phẩm bị lặp trong chương trình khuyến mãi",
	"promotion_items_product_id_fkey": "Sản phẩm không tồn tại",
}

// mapErr chuyển lỗi pgx/Postgres sang lỗi domain (không lộ text SQL).
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "23503":
			if msg, ok := conflictMessages[pg.ConstraintName]; ok {
				return domain.Conflict(msg)
			}
			if pg.Code == "23503" {
				return domain.Conflict("Dữ liệu đang được tham chiếu hoặc không tồn tại")
			}
			return domain.Conflict("Dữ liệu bị trùng")
		case "23514", "22001", "22P02", "22003", "22007", "22008":
			return domain.BadRequest("Dữ liệu không hợp lệ")
		}
	}
	return err
}

// likePattern thoát ký tự đặc biệt của LIKE và bọc %...%.
func likePattern(q string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(q) + "%"
}

func collect[T any](rows pgx.Rows, scan func(pgx.Rows) (T, error)) ([]T, error) {
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
