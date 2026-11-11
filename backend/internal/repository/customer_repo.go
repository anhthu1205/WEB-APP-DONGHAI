package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

const customerCols = `c.id, c.full_name, c.phone, c.address, c.note, c.created_by, c.created_at, c.updated_at,
	(SELECT count(*) FROM orders o WHERE o.customer_id = c.id)::int`

// Phạm vi dữ liệu theo người nhập: owner = nil → admin, xem tất cả; owner = &userID → nhân viên,
// chỉ thấy khách do mình tạo hoặc có ít nhất 1 đơn do mình nhập, và chỉ đếm đơn của mình.

// customerColsScoped giống customerCols nhưng orders_count chỉ đếm đơn trong phạm vi tham số $n.
func customerColsScoped(n int) string {
	p := fmt.Sprintf("$%d::bigint", n)
	return `c.id, c.full_name, c.phone, c.address, c.note, c.created_by, c.created_at, c.updated_at,
	(SELECT count(*) FROM orders o WHERE o.customer_id = c.id AND (` + p + ` IS NULL OR o.created_by = ` + p + `))::int`
}

// customerVisible là điều kiện khách nằm trong phạm vi tham số $n.
func customerVisible(n int) string {
	p := fmt.Sprintf("$%d::bigint", n)
	return `(` + p + ` IS NULL OR c.created_by = ` + p + `
	OR EXISTS (SELECT 1 FROM orders vo WHERE vo.customer_id = c.id AND vo.created_by = ` + p + `))`
}

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

// ListCustomers tìm khách trong phạm vi owner, có phân trang (mới nhất trước).
func (s *Store) ListCustomers(ctx context.Context, f CustomerSearch, owner *int64, pp domain.PageParams) (*domain.Page[*domain.Customer], error) {
	args := []any{f.Name, f.Digits, f.Phone, likePattern(f.Name), likePattern(f.Digits), owner}
	where := `WHERE (` + customerWhere[len("WHERE "):] + `) AND ` + customerVisible(6)
	var total int
	if err := s.q.QueryRow(ctx, `SELECT count(*) FROM customers c `+where, args...).Scan(&total); err != nil {
		return nil, mapErr(err)
	}
	rows, err := s.q.Query(ctx, `SELECT `+customerColsScoped(6)+` FROM customers c `+where+`
		ORDER BY c.id DESC LIMIT $7 OFFSET $8`, append(args, pp.PageSize, pp.Offset())...)
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

// GetCustomerScoped lấy khách theo id trong phạm vi owner (ngoài phạm vi → ErrNotFound).
func (s *Store) GetCustomerScoped(ctx context.Context, id int64, owner *int64) (*domain.Customer, error) {
	return scanCustomer(s.q.QueryRow(ctx, `SELECT `+customerColsScoped(2)+` FROM customers c
		WHERE c.id = $1 AND `+customerVisible(2), id, owner))
}

// CustomerInScope cho biết khách id có nằm trong phạm vi owner không.
func (s *Store) CustomerInScope(ctx context.Context, id int64, owner *int64) (bool, error) {
	var ok bool
	err := s.q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM customers c WHERE c.id = $1 AND `+customerVisible(2)+`)`, id, owner).Scan(&ok)
	return ok, mapErr(err)
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

// CountCustomers đếm số khách trong phạm vi owner.
func (s *Store) CountCustomers(ctx context.Context, owner *int64) (int, error) {
	var n int
	err := s.q.QueryRow(ctx, `SELECT count(*) FROM customers c WHERE `+customerVisible(1), owner).Scan(&n)
	return n, mapErr(err)
}
