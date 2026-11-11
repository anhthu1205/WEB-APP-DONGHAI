package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"donghai/internal/domain"
	"donghai/internal/phone"
	"donghai/internal/repository"
)

// CustomerRequest là body POST /api/internal/customers và PUT /api/admin/customers/{id}.
type CustomerRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Note     string `json:"note"`
}

// CustomerDetail là khách kèm danh sách đơn.
type CustomerDetail struct {
	Customer *domain.Customer
	Orders   []*domain.Order
}

// customerSearch dựng điều kiện tìm từ chuỗi q: SĐT chuẩn hóa được → khớp chính xác;
// có ≥ 3 chữ số → khớp một phần SĐT; luôn thử khớp tên.
func customerSearch(q string) repository.CustomerSearch {
	q = strings.TrimSpace(q)
	if q == "" {
		return repository.CustomerSearch{}
	}
	cs := repository.CustomerSearch{Name: q}
	if p, err := phone.Normalize(q); err == nil {
		cs.Phone = p
	}
	var digits strings.Builder
	for _, r := range q {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	if digits.Len() >= 3 {
		cs.Digits = digits.String()
	}
	return cs
}

// DataScope trả phạm vi dữ liệu của người dùng: admin → nil (xem tất cả);
// nhân viên → id của chính họ (chỉ dữ liệu do họ nhập).
func DataScope(me *domain.User) *int64 {
	if me.IsAdmin() {
		return nil
	}
	id := me.ID
	return &id
}

// ListCustomers tìm khách trong phạm vi của me.
func (s *Service) ListCustomers(ctx context.Context, me *domain.User, q string, pp domain.PageParams) (*domain.Page[*domain.Customer], error) {
	return s.store.ListCustomers(ctx, customerSearch(q), DataScope(me), pp)
}

// GetCustomerByPhone tìm chính xác theo SĐT chuẩn hóa, trong phạm vi của me
// (khách ngoài phạm vi trả 404 như không tồn tại, để không lộ tên khách của người khác).
func (s *Service) GetCustomerByPhone(ctx context.Context, me *domain.User, raw string) (*domain.Customer, error) {
	p, err := phone.Normalize(raw)
	if err != nil {
		return nil, &domain.ValidationError{Fields: map[string]string{"phone": "Số điện thoại không hợp lệ"}}
	}
	const msg = "Không tìm thấy khách hàng với số điện thoại này"
	c, err := s.store.GetCustomerByPhone(ctx, p)
	if err != nil {
		return nil, notFound(err, msg)
	}
	if scope := DataScope(me); scope != nil {
		c, err = s.store.GetCustomerScoped(ctx, c.ID, scope)
		return c, notFound(err, msg)
	}
	return c, nil
}

// GetCustomerDetail: khách + đơn hàng trong phạm vi của me.
func (s *Service) GetCustomerDetail(ctx context.Context, me *domain.User, id int64) (*CustomerDetail, error) {
	scope := DataScope(me)
	c, err := s.store.GetCustomerScoped(ctx, id, scope)
	if err != nil {
		return nil, notFound(err, "Không tìm thấy khách hàng")
	}
	orders, err := s.store.ListOrdersByCustomer(ctx, id, scope)
	if err != nil {
		return nil, err
	}
	return &CustomerDetail{Customer: c, Orders: orders}, nil
}

func buildCustomer(in CustomerRequest) (*domain.Customer, error) {
	c := &domain.Customer{
		FullName: strings.Join(strings.Fields(in.FullName), " "),
		Address:  strings.TrimSpace(in.Address),
		Note:     strings.TrimSpace(in.Note),
	}
	var v domain.Validator
	v.Check(c.FullName != "" && runeLen(c.FullName) <= 100, "full_name", "Họ tên bắt buộc, tối đa 100 ký tự")
	p, err := phone.Normalize(in.Phone)
	v.Check(err == nil, "phone", "Số điện thoại không hợp lệ")
	c.Phone = p
	v.Check(runeLen(c.Address) <= 300, "address", "Địa chỉ tối đa 300 ký tự")
	v.Check(runeLen(c.Note) <= 1000, "note", "Ghi chú tối đa 1000 ký tự")
	return c, v.Err()
}

func phoneConflict(existing *domain.Customer) error {
	return &domain.Error{Kind: domain.ErrConflict, Message: "Số điện thoại đã thuộc về khách hàng " + existing.FullName, ExistingID: existing.ID}
}

// CreateCustomer thêm khách (user + admin). Trùng SĐT → 409 kèm existing_id.
func (s *Service) CreateCustomer(ctx context.Context, me *domain.User, in CustomerRequest) (*domain.Customer, error) {
	c, err := buildCustomer(in)
	if err != nil {
		return nil, err
	}
	if ex, err := s.store.GetCustomerByPhone(ctx, c.Phone); err == nil {
		visible, err := s.store.CustomerInScope(ctx, ex.ID, DataScope(me))
		if err != nil {
			return nil, err
		}
		if !visible {
			// Khách của nhân viên khác: không lộ tên/id.
			return nil, domain.Conflict("Số điện thoại này đã có trong hệ thống. Hãy vào Nhập đơn hàng và nhập số này để tạo đơn cho khách.")
		}
		return nil, phoneConflict(ex)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	c.CreatedBy = &me.ID
	id, err := s.store.InsertCustomer(ctx, c)
	if err != nil {
		return nil, err
	}
	return s.store.GetCustomer(ctx, id)
}

// UpdateCustomer sửa khách (admin). Đổi SĐT trùng khách khác → 409.
func (s *Service) UpdateCustomer(ctx context.Context, id int64, in CustomerRequest) (*domain.Customer, error) {
	c, err := buildCustomer(in)
	if err != nil {
		return nil, err
	}
	c.ID = id
	if ex, err := s.store.GetCustomerByPhone(ctx, c.Phone); err == nil && ex.ID != id {
		return nil, phoneConflict(ex)
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if err := s.store.UpdateCustomer(ctx, c); err != nil {
		return nil, notFound(err, "Không tìm thấy khách hàng")
	}
	return s.store.GetCustomer(ctx, id)
}

// DeleteCustomer xóa khách chưa có đơn (admin).
func (s *Service) DeleteCustomer(ctx context.Context, id int64) error {
	c, err := s.store.GetCustomer(ctx, id)
	if err != nil {
		return notFound(err, "Không tìm thấy khách hàng")
	}
	if c.OrdersCount > 0 {
		return domain.Conflict("Khách hàng đã có đơn hàng, không thể xóa")
	}
	return notFound(s.store.DeleteCustomer(ctx, id), "Không tìm thấy khách hàng")
}
