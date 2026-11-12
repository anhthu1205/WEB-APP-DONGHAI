package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

// getImage phục vụ bytes ảnh với cache immutable + ETag, hỗ trợ 304.
func (h *Handler) getImage(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(chi.URLParam(r, "uuid"))
	if !service.ValidUUID(id) {
		mw.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy ảnh")
		return
	}
	etag := `"` + id + `"`
	if inm := r.Header.Get("If-None-Match"); inm != "" {
		for _, t := range strings.Split(inm, ",") {
			t = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), "W/"))
			if t == etag || t == "*" {
				w.Header().Set("ETag", etag)
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}
	mime, data, err := h.svc.GetImage(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	hd := w.Header()
	hd.Set("Content-Type", mime)
	hd.Set("Content-Length", strconv.Itoa(len(data)))
	hd.Set("Cache-Control", "public, max-age=31536000, immutable")
	hd.Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

// readUploadFile đọc trường multipart "file" (≤ service.MaxImageBytes). Trả false nếu đã ghi lỗi ra w.
func readUploadFile(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxImageBytes+64<<10)
	mr, err := r.MultipartReader()
	if err != nil {
		mw.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Yêu cầu phải là multipart/form-data với trường file")
		return nil, false
	}
	var data []byte
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			uploadReadError(w, err)
			return nil, false
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		data, err = io.ReadAll(io.LimitReader(part, service.MaxImageBytes+1))
		_ = part.Close()
		if err != nil {
			uploadReadError(w, err)
			return nil, false
		}
		if len(data) > service.MaxImageBytes {
			mw.WriteError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Ảnh vượt quá 2 MB")
			return nil, false
		}
		break
	}
	if data == nil {
		writeJSON(w, http.StatusBadRequest, mw.ErrorBody{Error: mw.ErrorDetail{
			Code: "VALIDATION_ERROR", Message: "Dữ liệu không hợp lệ", Fields: map[string]string{"file": "Vui lòng chọn ảnh"},
		}})
		return nil, false
	}
	return data, true
}

// uploadImage nhận multipart field "file" (1 ảnh/lần, ≤ 2 MB).
func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request) {
	pid, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	data, ok := readUploadFile(w, r)
	if !ok {
		return
	}
	im, err := h.svc.UploadImage(r.Context(), pid, data)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toInternalImage(im))
}

// getLogo phục vụ logo. URL có ?v=<version> khớp → cache 1 năm (immutable); không có v → cache ngắn.
func (h *Handler) getLogo(w http.ResponseWriter, r *http.Request) {
	mime, ver, data, err := h.svc.GetLogo(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	etag := `"logo-` + ver + `"`
	cache := "public, max-age=300"
	if r.URL.Query().Get("v") == ver {
		cache = "public, max-age=31536000, immutable"
	}
	hd := w.Header()
	hd.Set("Cache-Control", cache)
	hd.Set("ETag", etag)
	if inm := r.Header.Get("If-None-Match"); inm != "" {
		for _, t := range strings.Split(inm, ",") {
			if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), "W/")) == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}
	hd.Set("Content-Type", mime)
	hd.Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

// uploadLogo thay logo công ty (admin).
func (h *Handler) uploadLogo(w http.ResponseWriter, r *http.Request) {
	data, ok := readUploadFile(w, r)
	if !ok {
		return
	}
	c, err := h.svc.UploadLogo(r.Context(), mw.UserFromContext(r.Context()), data)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPublicCompany(c))
}

// deleteLogo xóa logo riêng, quay về logo mặc định (admin).
func (h *Handler) deleteLogo(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.DeleteLogo(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPublicCompany(c))
}

func uploadReadError(w http.ResponseWriter, err error) {
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) || errors.Is(err, multipart.ErrMessageTooLarge) {
		mw.WriteError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Ảnh vượt quá 2 MB")
		return
	}
	mw.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Không đọc được tệp tải lên")
}

func (h *Handler) setMainImage(w http.ResponseWriter, r *http.Request) {
	var in service.SetMainRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	im, err := h.svc.SetMainImage(r.Context(), strings.ToLower(chi.URLParam(r, "uuid")), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toInternalImage(im))
}

func (h *Handler) deleteImage(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteImage(r.Context(), strings.ToLower(chi.URLParam(r, "uuid"))); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
