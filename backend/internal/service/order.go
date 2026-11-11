package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"donghai/internal/domain"
	"donghai/internal/phone"
	"donghai/internal/repository"
)

// Giới hạn số dòng của đơn.
const (
	MinOrderItems = 1
	MaxOrderItems = 50
)

// OrderCustomerRequest là thông tin khách nhập kèm đơn.
type OrderCustomerRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

// OrderItemRequest là 1 dòng SP của đơn; unit_price bỏ trống → giá niêm yết.
type OrderItemRequest struct {
	ProductID    int64   `json:"product_id"`
	Quantity     int     `json:"quantity"`
	UnitPrice    *int64  `json:"unit_price"`
	SerialNumber *string `json:"serial_number"`
}

// CreateOrderRequest là body POST /api/internal/orders: customer_id HOẶC customer.
type CreateOrderRequest struct {
	CustomerID   *int64                `json:"customer_id"`
	Customer     *OrderCustomerRequest `json:"customer"`
	PurchaseDate string                `json:"purchase_date"`
	Note         string                `json:"note"`
	Items        []OrderItemRequest    `json:"items"`
}

// UpdateOrderRequest là body PUT /api/admin/orders/{id} (không có customer).
type UpdateOrderRequest struct {
	CustomerID   *int64             `json:"customer_id"`
	PurchaseDate string             `json:"purchase_date"`
	Note         string             `json:"note"`
	Items        []OrderItemRequest `json:"items"`
}

// OrderFilterRequest là tham số GET /api/internal/orders.
type OrderFilterRequest struct {
	Q    string
	From string
	To   string
}

// Dashboard là số liệu tổng quan nội bộ.
type Dashboard struct {
	CustomersTotal  int
	OrdersThisMonth int
	OrdersTotal     int
	RecentOrders    []*domain.Order
}

// OrderCode = "DH" + id 6 chữ số.
func OrderCode(id int64) string { return fmt.Sprintf("DH%06d", id) }

// ListOrders tìm đơn.
func (s *Service) ListOrders(ctx context.Context, f OrderFilterRequest, pp domain.PageParams) (*domain.Page[*domain.Order], error) {
	var v domain.Validator
	of := repository.OrderFilter{Customer: customerSearch(f.Q)}
	if strings.TrimSpace(f.From) != "" {
		t := parseDate(&v, "from", f.From)
		of.From = &t
	}
	if strings.TrimSpace(f.To) != "" {
		t := parseDate(&v, "to", f.To)
		of.To = &t
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	return s.store.ListOrders(ctx, of, pp)
}

// GetOrder chi tiết đơn.
func (s *Service) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	o, err := s.store.GetOrder(ctx, id)
	return o, notFound(err, "Không tìm thấy đơn hàng")
}

// GetDashboard số liệu tổng quan.
func (s *Service) GetDashboard(ctx context.Context) (*Dashboard, error) {
	today := s.Today()
	first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	d := &Dashboard{}
	var err error
	if d.CustomersTotal, err = s.store.CountCustomers(ctx); err != nil {
		return nil, err
	}
	if d.OrdersTotal, d.OrdersThisMonth, err = s.store.OrderStats(ctx, first, last); err != nil {
		return nil, err
	}
	if d.RecentOrders, err = s.store.RecentOrders(ctx, 5); err != nil {
		return nil, err
	}
	return d, nil
}

// validateOrderBase kiểm tra ngày mua, ghi chú, dòng SP (chưa tra SP trong DB).
func (s *Service) validateOrderBase(v *domain.Validator, date, note string, items []OrderItemRequest) (time.Time, string) {
	d := parseDate(v, "purchase_date", date)
	if !d.IsZero() {
		if d.After(s.Today()) {
			v.Add("purchase_date", "Ngày mua không được ở tương lai")
		} else if d.Before(minPurchaseDate) {
			v.Add("purchase_date", "Ngày mua không được trước 01/01/2000")
		}
	}
	note = strings.TrimSpace(note)
	v.Check(runeLen(note) <= 1000, "note", "Ghi chú tối đa 1000 ký tự")
	if len(items) < MinOrderItems || len(items) > MaxOrderItems {
		v.Add("items", "Đơn hàng phải có từ 1 đến 50 dòng sản phẩm")
	}
	for i, it := range items {
		v.Check(it.ProductID > 0, fmt.Sprintf("items.%d.product_id", i), "Vui lòng chọn sản phẩm")
		v.Check(it.Quantity >= 1 && it.Quantity <= 10000, fmt.Sprintf("items.%d.quantity", i), "Số lượng từ 1 đến 10000")
		if it.UnitPrice != nil {
			v.Check(*it.UnitPrice >= 0 && *it.UnitPrice <= MaxPrice, fmt.Sprintf("items.%d.unit_price", i), "Đơn giá không hợp lệ")
		}
		if sn := trimPtr(it.SerialNumber); sn != nil {
			v.Check(runeLen(*sn) <= 100, fmt.Sprintf("items.%d.serial_number", i), "Số serial tối đa 100 ký tự")
		}
	}
	return d, note
}

// resolveItems tra SP trong DB, chụp tên + giá mặc định.
func resolveItems(ctx context.Context, tx *serviceTx, items []OrderItemRequest) ([]domain.OrderItem, error) {
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ProductID)
	}
	m, err := tx.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	var v domain.Validator
	out := make([]domain.OrderItem, 0, len(items))
	for i, it := range items {
		p := m[it.ProductID]
		if p == nil {
			v.Add(fmt.Sprintf("items.%d.product_id", i), "Sản phẩm không tồn tại")
			continue
		}
		price := p.PricePublic
		if it.UnitPrice != nil {
			price = *it.UnitPrice
		}
		out = append(out, domain.OrderItem{
			ProductID: p.ID, ProductName: p.Name, Quantity: it.Quantity, UnitPrice: price,
			SerialNumber: trimPtr(it.SerialNumber),
		})
	}
	return out, v.Err()
}

// CreateOrder tạo đơn (+ khách mới nếu cần) trong 1 transaction. Trả đơn và cờ dùng lại khách.
func (s *Service) CreateOrder(ctx context.Context, me *domain.User, in CreateOrderRequest) (*domain.Order, bool, error) {
	var v domain.Validator
	date, note := s.validateOrderBase(&v, in.PurchaseDate, in.Note, in.Items)

	var newCustomer *domain.Customer
	switch {
	case in.CustomerID != nil && in.Customer != nil:
		v.Add("customer", "Chỉ gửi customer_id hoặc customer, không gửi cả hai")
	case in.CustomerID != nil:
		v.Check(*in.CustomerID > 0, "customer_id", "Khách hàng không hợp lệ")
	case in.Customer != nil:
		c, err := buildCustomer(CustomerRequest{FullName: in.Customer.FullName, Phone: in.Customer.Phone, Address: in.Customer.Address})
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			for k, msg := range ve.Fields {
				// SĐT hợp lệ + khách đã tồn tại thì không cần họ tên → kiểm tra lại sau
				v.Add("customer."+k, msg)
			}
		}
		newCustomer = c
	default:
		v.Add("customer", "Vui lòng chọn hoặc nhập thông tin khách hàng")
	}
	// Nếu SĐT đã có khách thì cho phép bỏ trống họ tên.
	verr := v.Err()
	if verr != nil && newCustomer != nil && newCustomer.Phone != "" && newCustomer.FullName == "" {
		if _, err := s.store.GetCustomerByPhone(ctx, newCustomer.Phone); err == nil {
			if ve, ok := verr.(*domain.ValidationError); ok {
				delete(ve.Fields, "customer.full_name")
				if len(ve.Fields) == 0 {
					verr = nil
				}
			}
		}
	}
	if verr != nil {
		return nil, false, verr
	}

	var orderID int64
	reused := false
	err := s.store.InTx(ctx, func(tx *serviceTx) error {
		var customerID int64
		if in.CustomerID != nil {
			if _, err := tx.GetCustomer(ctx, *in.CustomerID); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return &domain.ValidationError{Fields: map[string]string{"customer_id": "Khách hàng không tồn tại"}}
				}
				return err
			}
			customerID = *in.CustomerID
		} else {
			ex, err := tx.GetCustomerByPhone(ctx, newCustomer.Phone)
			switch {
			case err == nil:
				customerID, reused = ex.ID, true // không ghi đè tên/địa chỉ
			case errors.Is(err, domain.ErrNotFound):
				newCustomer.CreatedBy = &me.ID
				if customerID, err = tx.InsertCustomer(ctx, newCustomer); err != nil {
					return err
				}
			default:
				return err
			}
		}
		items, err := resolveItems(ctx, tx, in.Items)
		if err != nil {
			return err
		}
		orderID, err = tx.InsertOrder(ctx, &domain.Order{CustomerID: customerID, PurchaseDate: date, Note: note, CreatedBy: &me.ID})
		if err != nil {
			return err
		}
		return tx.ReplaceOrderItems(ctx, orderID, items)
	})
	if err != nil {
		return nil, false, err
	}
	o, err := s.GetOrder(ctx, orderID)
	return o, reused, err
}

// UpdateOrder sửa đơn và thay items (admin).
func (s *Service) UpdateOrder(ctx context.Context, id int64, in UpdateOrderRequest) (*domain.Order, error) {
	var v domain.Validator
	date, note := s.validateOrderBase(&v, in.PurchaseDate, in.Note, in.Items)
	v.Check(in.CustomerID != nil && *in.CustomerID > 0, "customer_id", "Vui lòng chọn khách hàng")
	if err := v.Err(); err != nil {
		return nil, err
	}
	err := s.store.InTx(ctx, func(tx *serviceTx) error {
		if _, err := tx.GetCustomer(ctx, *in.CustomerID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return &domain.ValidationError{Fields: map[string]string{"customer_id": "Khách hàng không tồn tại"}}
			}
			return err
		}
		items, err := resolveItems(ctx, tx, in.Items)
		if err != nil {
			return err
		}
		if err := tx.UpdateOrder(ctx, &domain.Order{ID: id, CustomerID: *in.CustomerID, PurchaseDate: date, Note: note}); err != nil {
			return notFound(err, "Không tìm thấy đơn hàng")
		}
		return tx.ReplaceOrderItems(ctx, id, items)
	})
	if err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, id)
}

// DeleteOrder xóa đơn (admin).
func (s *Service) DeleteOrder(ctx context.Context, id int64) error {
	return notFound(s.store.DeleteOrder(ctx, id), "Không tìm thấy đơn hàng")
}

// normalizePhoneField chuẩn hóa SĐT, lỗi → ValidationError trường phone.
func normalizePhoneField(raw string) (string, error) {
	p, err := phone.Normalize(raw)
	if err != nil {
		return "", &domain.ValidationError{Fields: map[string]string{"phone": "Số điện thoại không hợp lệ"}}
	}
	return p, nil
}
