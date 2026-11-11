package service

import (
	"context"
	"errors"
	"time"

	"donghai/internal/domain"
	"donghai/internal/phone"
)

// WarrantyLookupRequest là body POST /api/public/warranty/lookup.
type WarrantyLookupRequest struct {
	Phone string `json:"phone"`
}

// WarrantyLookupResult: dữ liệu đã che, không có địa chỉ/giá/ghi chú/người nhập.
type WarrantyLookupResult struct {
	PhoneMasked        string
	CustomerNameMasked string
	Today              time.Time
	Orders             []*domain.Order
}

// LookupWarranty tra đơn theo SĐT. SĐT hợp lệ nhưng không có khách → orders rỗng.
func (s *Service) LookupWarranty(ctx context.Context, in WarrantyLookupRequest) (*WarrantyLookupResult, error) {
	p, err := normalizePhoneField(in.Phone)
	if err != nil {
		return nil, err
	}
	res := &WarrantyLookupResult{PhoneMasked: phone.Mask(p), Today: s.Today(), Orders: []*domain.Order{}}
	c, err := s.store.GetCustomerByPhone(ctx, p)
	if errors.Is(err, domain.ErrNotFound) {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	res.CustomerNameMasked = phone.MaskName(c.FullName)
	if res.Orders, err = s.store.ListOrdersByCustomer(ctx, c.ID); err != nil {
		return nil, err
	}
	return res, nil
}
