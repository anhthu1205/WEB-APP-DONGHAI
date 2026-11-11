package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"donghai/internal/domain"
	mw "donghai/internal/middleware"
)

// MaxJSONBody là giới hạn body JSON (1 MB).
const MaxJSONBody = 1 << 20

func writeJSON(w http.ResponseWriter, status int, v any) { mw.WriteJSON(w, status, v) }

// writeError map lỗi domain → HTTP. Lỗi lạ → 500 message chung, log kèm request_id.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		writeJSON(w, http.StatusBadRequest, mw.ErrorBody{Error: mw.ErrorDetail{
			Code: "VALIDATION_ERROR", Message: "Dữ liệu không hợp lệ", Fields: ve.Fields,
		}})
		return
	}
	var de *domain.Error
	msg := ""
	var existing int64
	if errors.As(err, &de) {
		msg, existing = de.Message, de.ExistingID
	}
	type m struct {
		kind   error
		status int
		code   string
		def    string
	}
	for _, x := range []m{
		{domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy"},
		{domain.ErrConflict, http.StatusConflict, "CONFLICT", "Dữ liệu bị trùng hoặc đang được sử dụng"},
		{domain.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED", "Vui lòng đăng nhập"},
		{domain.ErrForbidden, http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền thực hiện thao tác này"},
		{domain.ErrBadRequest, http.StatusBadRequest, "BAD_REQUEST", "Yêu cầu không hợp lệ"},
		{domain.ErrPayloadTooLarge, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Dữ liệu quá lớn"},
		{domain.ErrUnsupportedMedia, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Định dạng không được hỗ trợ"},
	} {
		if errors.Is(err, x.kind) {
			if msg == "" {
				msg = x.def
			}
			writeJSON(w, x.status, mw.ErrorBody{Error: mw.ErrorDetail{Code: x.code, Message: msg, ExistingID: existing}})
			return
		}
	}
	slog.ErrorContext(r.Context(), "lỗi nội bộ", "request_id", mw.RequestIDFrom(r.Context()),
		"method", r.Method, "path", r.URL.Path, "error", err.Error())
	mw.WriteError(w, http.StatusInternalServerError, "INTERNAL", "Đã có lỗi xảy ra")
}

// decodeJSON đọc body JSON: ≤ 1 MB, cấm trường lạ, chỉ 1 giá trị JSON.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err == nil {
		if dec.Decode(&struct{}{}) != io.EOF {
			err = errors.New("dư dữ liệu sau JSON")
		}
	}
	if err == nil {
		return true
	}
	var mbe *http.MaxBytesError
	switch {
	case errors.As(err, &mbe):
		mw.WriteError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Dữ liệu gửi lên vượt quá 1 MB")
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		field := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
		mw.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Trường không được phép: "+field)
	case errors.Is(err, io.EOF):
		mw.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Thiếu dữ liệu JSON")
	default:
		var ute *json.UnmarshalTypeError
		if errors.As(err, &ute) && ute.Field != "" {
			writeJSON(w, http.StatusBadRequest, mw.ErrorBody{Error: mw.ErrorDetail{
				Code: "VALIDATION_ERROR", Message: "Dữ liệu không hợp lệ",
				Fields: map[string]string{ute.Field: "Sai kiểu dữ liệu"},
			}})
			return false
		}
		mw.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "JSON không hợp lệ")
	}
	return false
}

// idParam đọc {id} dạng số dương; sai → 404.
func idParam(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		mw.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy")
		return 0, false
	}
	return id, true
}

func pageParams(r *http.Request) domain.PageParams {
	q := r.URL.Query()
	p, _ := strconv.Atoi(q.Get("page"))
	s, _ := strconv.Atoi(q.Get("page_size"))
	return domain.NewPageParams(p, s)
}

// PageDTO là định dạng phân trang chung.
type PageDTO[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func toPage[S any, T any](p *domain.Page[S], f func(S) T) PageDTO[T] {
	out := PageDTO[T]{Items: make([]T, 0, len(p.Items)), Total: p.Total, Page: p.Page, PageSize: p.PageSize}
	for _, it := range p.Items {
		out.Items = append(out.Items, f(it))
	}
	return out
}
