package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

const customerCols = `c.id, c.full_name, c.phone, c.address, c.note, c.created_by, c.created_at, c.updated_at,
	(SELECT count(*) FROM orders o WHERE o.customer_id = c.id)::int`

func scanCustomer(r pgx.Row) (*domain.Customer, error) {
	var c domain.Customer
	err := r.Scan(&c.ID, &c.FullName, &c.Phone, &c.Address, &c.Note, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.OrdersCount)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

// CustomerSearch: Name khớp ILIKE tên; Digits khớp một phần SĐT; Phone khớp chính xác.
type CustomerSearch struct {
	Name   string
	Digits string
	Phone  string
}

const customerWhere = `WHERE ($1 = '' AND $2 = '' AND $3 = '')
	OR ($1 <> '' AND c.full_name ILIKE $4)
	OR ($2 <> '' AND c.phone LIKE $5)
	OR ($3 <> '' AND c.phone = $3)`

// ListCustomers tìm khách có phân trang (mới nhất trước).
func (s *Store) ListCustomers(ctx context.Context, f CustomerSearch, pp domain.PageParams) (*domain.Page[*domain.Customer], error) {
	args := []any{f.Name, f.Digits, f.Phone, likePattern(f.Name), likePattern(f.Digits)}
	var total int
	if err := s.q.QueryRow(ctx, `SELECT count(*) FROM customers c `+customerWhere, args...).Scan(&total); err != nil {
		return nil, mapErr(err)
	}
	rows, err := s.q.Query(ctx, `SELECT `+customerCols+` FROM customers c `+customerWhere+`
		ORDER BY c.id DESC LIMIT $6 OFFSET $7`, append(args, pp.PageSize, pp.Offset())...)
	if err != nil {
		return nil, mapErr(err)
	}
	items, err := collect(rows, func(r pgx.Rows) (*domain.Customer, error) { return scanCustomer(r) })
	if err != nil {
		return nil, err
	}
	return &domain.Page[*domain.Customer]{Items: items, Total: total, Page: pp.Page, PageSize: pp.PageSize}, nil
}

// GetCustomer lấy khách theo id.
func (s *Store) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	return scanCustomer(s.q.QueryRow(ctx, `SELECT `+customerCols+` FROM customers c WHERE c.id = $1`, id))
}

// GetCustomerByPhone lấy khách theo SĐT đã chuẩn hóa.
func (s *Store) GetCustomerByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	return scanCustomer(s.q.QueryRow(ctx, `SELECT `+customerCols+` FROM customers c WHERE c.phone = $1`, phone))
}

// InsertCustomer thêm khách, trả id.
func (s *Store) InsertCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	var id int64
	err := s.q.QueryRow(ctx, `INSERT INTO customers (full_name, phone, address, note, created_by)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`, c.FullName, c.Phone, c.Address, c.Note, c.CreatedBy).Scan(&id)
	return id, mapErr(err)
}

// UpdateCustomer sửa khách.
func (s *Store) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	tag, err := s.q.Exec(ctx, `UPDATE customers SET full_name = $2, phone = $3, address = $4, note = $5 WHERE id = $1`,
		c.ID, c.FullName, c.Phone, c.Address, c.Note)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteCustomer xóa khách (có đơn → lỗi FK → 409).
func (s *Store) DeleteCustomer(ctx context.Context, id int64) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CountCustomers đếm tổng số khách.
func (s *Store) CountCustomers(ctx context.Context) (int, error) {
	var n int
	err := s.q.QueryRow(ctx, `SELECT count(*) FROM customers`).Scan(&n)
	return n, mapErr(err)
}
