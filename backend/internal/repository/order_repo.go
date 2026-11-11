package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"donghai/internal/domain"
)

const orderCols = `o.id, o.customer_id, o.purchase_date, o.note, o.created_by, u.username, o.created_at, o.updated_at,
	c.id, c.full_name, c.phone, c.address,
	COALESCE((SELECT sum(oi.quantity::bigint * oi.unit_price) FROM order_items oi WHERE oi.order_id = o.id), 0)::bigint,
	COALESCE((SELECT sum(oi.quantity) FROM order_items oi WHERE oi.order_id = o.id), 0)::int`

const orderFrom = ` FROM orders o JOIN customers c ON c.id = o.customer_id LEFT JOIN users u ON u.id = o.created_by `

func scanOrder(r pgx.Row) (*domain.Order, error) {
	var o domain.Order
	var c domain.Customer
	err := r.Scan(&o.ID, &o.CustomerID, &o.PurchaseDate, &o.Note, &o.CreatedBy, &o.CreatedByUsername,
		&o.CreatedAt, &o.UpdatedAt, &c.ID, &c.FullName, &c.Phone, &c.Address, &o.TotalAmount, &o.ItemsCount)
	if err != nil {
		return nil, mapErr(err)
	}
	o.Customer = &c
	return &o, nil
}

// OrderFilter lọc đơn hàng.
type OrderFilter struct {
	Customer CustomerSearch
	From     *time.Time
	To       *time.Time
}

const orderWhere = `WHERE (($1 = '' AND $2 = '' AND $3 = '')
	OR ($1 <> '' AND c.full_name ILIKE $4)
	OR ($2 <> '' AND c.phone LIKE $5)
	OR ($3 <> '' AND c.phone = $3))
	AND ($6::date IS NULL OR o.purchase_date >= $6::date)
	AND ($7::date IS NULL OR o.purchase_date <= $7::date)`

// ListOrders tìm đơn có phân trang (ngày mua mới nhất trước).
func (s *Store) ListOrders(ctx context.Context, f OrderFilter, pp domain.PageParams) (*domain.Page[*domain.Order], error) {
	cs := f.Customer
	args := []any{cs.Name, cs.Digits, cs.Phone, likePattern(cs.Name), likePattern(cs.Digits), f.From, f.To}
	var total int
	if err := s.q.QueryRow(ctx, `SELECT count(*)`+orderFrom+orderWhere, args...).Scan(&total); err != nil {
		return nil, mapErr(err)
	}
	rows, err := s.q.Query(ctx, `SELECT `+orderCols+orderFrom+orderWhere+`
		ORDER BY o.purchase_date DESC, o.id DESC LIMIT $8 OFFSET $9`, append(args, pp.PageSize, pp.Offset())...)
	if err != nil {
		return nil, mapErr(err)
	}
	items, err := collect(rows, func(r pgx.Rows) (*domain.Order, error) { return scanOrder(r) })
	if err != nil {
		return nil, err
	}
	return &domain.Page[*domain.Order]{Items: items, Total: total, Page: pp.Page, PageSize: pp.PageSize}, nil
}

// RecentOrders lấy n đơn mới nhập gần đây.
func (s *Store) RecentOrders(ctx context.Context, n int) ([]*domain.Order, error) {
	rows, err := s.q.Query(ctx, `SELECT `+orderCols+orderFrom+` ORDER BY o.created_at DESC, o.id DESC LIMIT $1`, n)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(r pgx.Rows) (*domain.Order, error) { return scanOrder(r) })
}

// GetOrder lấy đơn kèm items.
func (s *Store) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	o, err := scanOrder(s.q.QueryRow(ctx, `SELECT `+orderCols+orderFrom+` WHERE o.id = $1`, id))
	if err != nil {
		return nil, err
	}
	m, err := s.orderItems(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	o.Items = m[id]
	return o, nil
}

// ListOrdersByCustomer lấy mọi đơn của khách (ngày mua giảm dần) kèm items.
func (s *Store) ListOrdersByCustomer(ctx context.Context, customerID int64) ([]*domain.Order, error) {
	rows, err := s.q.Query(ctx, `SELECT `+orderCols+orderFrom+` WHERE o.customer_id = $1
		ORDER BY o.purchase_date DESC, o.id DESC`, customerID)
	if err != nil {
		return nil, mapErr(err)
	}
	list, err := collect(rows, func(r pgx.Rows) (*domain.Order, error) { return scanOrder(r) })
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(list))
	for i, o := range list {
		ids[i] = o.ID
	}
	m, err := s.orderItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, o := range list {
		o.Items = m[o.ID]
	}
	return list, nil
}

func (s *Store) orderItems(ctx context.Context, ids []int64) (map[int64][]domain.OrderItem, error) {
	out := map[int64][]domain.OrderItem{}
	for _, id := range ids {
		out[id] = []domain.OrderItem{}
	}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.q.Query(ctx, `SELECT order_id, id, product_id, product_name, quantity, unit_price, serial_number
		FROM order_items WHERE order_id = ANY($1) ORDER BY order_id, id`, ids)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var oid int64
		var it domain.OrderItem
		if err := rows.Scan(&oid, &it.ID, &it.ProductID, &it.ProductName, &it.Quantity, &it.UnitPrice, &it.SerialNumber); err != nil {
			return nil, err
		}
		out[oid] = append(out[oid], it)
	}
	return out, rows.Err()
}

// InsertOrder thêm đơn (không gồm items), trả id.
func (s *Store) InsertOrder(ctx context.Context, o *domain.Order) (int64, error) {
	var id int64
	err := s.q.QueryRow(ctx, `INSERT INTO orders (customer_id, purchase_date, note, created_by)
		VALUES ($1, $2, $3, $4) RETURNING id`, o.CustomerID, o.PurchaseDate, o.Note, o.CreatedBy).Scan(&id)
	return id, mapErr(err)
}

// UpdateOrder sửa đơn.
func (s *Store) UpdateOrder(ctx context.Context, o *domain.Order) error {
	tag, err := s.q.Exec(ctx, `UPDATE orders SET customer_id = $2, purchase_date = $3, note = $4 WHERE id = $1`,
		o.ID, o.CustomerID, o.PurchaseDate, o.Note)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ReplaceOrderItems thay toàn bộ dòng SP của đơn.
func (s *Store) ReplaceOrderItems(ctx context.Context, orderID int64, items []domain.OrderItem) error {
	if _, err := s.q.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderID); err != nil {
		return mapErr(err)
	}
	for _, it := range items {
		if _, err := s.q.Exec(ctx, `INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, serial_number)
			VALUES ($1, $2, $3, $4, $5, $6)`, orderID, it.ProductID, it.ProductName, it.Quantity, it.UnitPrice, it.SerialNumber); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

// DeleteOrder xóa đơn (items xóa theo cascade).
func (s *Store) DeleteOrder(ctx context.Context, id int64) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// OrderStats trả tổng số đơn và số đơn có ngày mua trong [from, to].
func (s *Store) OrderStats(ctx context.Context, from, to time.Time) (total, inRange int, err error) {
	err = s.q.QueryRow(ctx, `SELECT count(*), count(*) FILTER (WHERE purchase_date BETWEEN $1::date AND $2::date) FROM orders`,
		from, to).Scan(&total, &inRange)
	return total, inRange, mapErr(err)
}
