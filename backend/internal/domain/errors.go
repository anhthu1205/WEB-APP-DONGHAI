// Package domain chứa struct nghiệp vụ thuần và lỗi nghiệp vụ dùng chung.
package domain

import (
	"errors"
	"sort"
	"strings"
)

// Lỗi sentinel; handler map sang mã HTTP.
var (
	ErrNotFound         = errors.New("không tìm thấy")
	ErrConflict         = errors.New("xung đột dữ liệu")
	ErrUnauthorized     = errors.New("chưa xác thực")
	ErrForbidden        = errors.New("không có quyền")
	ErrBadRequest       = errors.New("yêu cầu không hợp lệ")
	ErrPayloadTooLarge  = errors.New("dữ liệu quá lớn")
	ErrUnsupportedMedia = errors.New("định dạng không được hỗ trợ")
)

// Error là lỗi nghiệp vụ có message cho người dùng, bọc một lỗi sentinel.
type Error struct {
	Kind       error
	Message    string
	ExistingID int64 // dùng cho 409 trùng SĐT khách hàng
}

func (e *Error) Error() string { return e.Message }

// Unwrap cho phép errors.Is(err, domain.ErrXxx).
func (e *Error) Unwrap() error { return e.Kind }

// NotFound tạo lỗi 404 với message tiếng Việt.
func NotFound(msg string) error { return &Error{Kind: ErrNotFound, Message: msg} }

// Conflict tạo lỗi 409.
func Conflict(msg string) error { return &Error{Kind: ErrConflict, Message: msg} }

// Unauthorized tạo lỗi 401.
func Unauthorized(msg string) error { return &Error{Kind: ErrUnauthorized, Message: msg} }

// BadRequest tạo lỗi 400 BAD_REQUEST.
func BadRequest(msg string) error { return &Error{Kind: ErrBadRequest, Message: msg} }

// ValidationError là lỗi 400 VALIDATION_ERROR theo từng trường.
type ValidationError struct {
	Fields map[string]string
}

func (v *ValidationError) Error() string {
	keys := make([]string, 0, len(v.Fields))
	for k := range v.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + ": " + v.Fields[k]
	}
	return "dữ liệu không hợp lệ: " + strings.Join(parts, "; ")
}

// Validator gom lỗi theo trường.
type Validator struct {
	fields map[string]string
}

// Add thêm lỗi cho trường (giữ lỗi đầu tiên).
func (v *Validator) Add(field, msg string) {
	if v.fields == nil {
		v.fields = map[string]string{}
	}
	if _, ok := v.fields[field]; !ok {
		v.fields[field] = msg
	}
}

// Check thêm lỗi khi cond sai.
func (v *Validator) Check(cond bool, field, msg string) {
	if !cond {
		v.Add(field, msg)
	}
}

// Err trả *ValidationError hoặc nil.
func (v *Validator) Err() error {
	if len(v.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: v.fields}
}
