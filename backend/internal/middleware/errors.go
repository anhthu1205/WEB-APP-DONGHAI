// Package middleware: xác thực, phân quyền, rate limit, CORS, header bảo mật, log.
package middleware

import (
	"encoding/json"
	"net/http"
)

// ErrorBody là định dạng lỗi chung {"error":{...}}.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail là nội dung lỗi.
type ErrorDetail struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Fields     map[string]string `json:"fields,omitempty"`
	ExistingID int64             `json:"existing_id,omitempty"`
}

// WriteJSON ghi JSON với status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// WriteError ghi lỗi theo định dạng chung.
func WriteError(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: msg}})
}
