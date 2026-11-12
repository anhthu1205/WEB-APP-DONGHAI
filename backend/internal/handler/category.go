package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"donghai/internal/domain"
	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

// PublicCategoryDTO là danh mục công khai (menu, trang danh mục).
type PublicCategoryDTO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Description  string  `json:"description"`
	ShowInMenu   bool    `json:"show_in_menu"`
	CoverURL     *string `json:"cover_url"`
	ProductCount int     `json:"product_count"`
}

// InternalCategoryDTO là danh mục cho trang quản trị.
type InternalCategoryDTO struct {
	PublicCategoryDTO
	IsVisible  bool      `json:"is_visible"`
	ShowOnHome bool      `json:"show_on_home"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CategoryPageDTO là trang danh mục: danh mục + SP.
type CategoryPageDTO struct {
	Category PublicCategoryDTO  `json:"category"`
	Products []PublicProductDTO `json:"products"`
}

func coverURL(c *domain.Category) *string {
	if c.CoverVersion == nil {
		return nil
	}
	u := "/api/public/category-covers/" + strconv.FormatInt(c.ID, 10) + "?v=" + *c.CoverVersion
	return &u
}

func toPublicCategory(c *domain.Category) PublicCategoryDTO {
	return PublicCategoryDTO{ID: c.ID, Name: c.Name, Slug: c.Slug, Description: c.Description,
		ShowInMenu: c.ShowInMenu, CoverURL: coverURL(c), ProductCount: c.ProductCount}
}

func toInternalCategory(c *domain.Category) InternalCategoryDTO {
	return InternalCategoryDTO{PublicCategoryDTO: toPublicCategory(c), IsVisible: c.IsVisible, ShowOnHome: c.ShowOnHome,
		SortOrder: c.SortOrder, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

func (h *Handler) publicCategories(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListPublicCategories(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]PublicCategoryDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toPublicCategory(c))
	}
	w.Header().Set("Cache-Control", "no-cache")
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) publicCategoryPage(w http.ResponseWriter, r *http.Request) {
	pg, err := h.svc.GetCategoryPage(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := CategoryPageDTO{Category: toPublicCategory(pg.Category), Products: make([]PublicProductDTO, 0, len(pg.Products))}
	for _, p := range pg.Products {
		out.Products = append(out.Products, toPublicProduct(p))
	}
	writeJSON(w, http.StatusOK, out)
}

// getCategoryCover phục vụ ảnh bìa danh mục (?v=<version> khớp → cache 1 năm, không có v → 5 phút; ETag/304).
func (h *Handler) getCategoryCover(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		mw.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy ảnh")
		return
	}
	mime, ver, data, err := h.svc.GetCategoryCover(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	etag := `"cat-` + ver + `"`
	cache := "public, max-age=300"
	if r.URL.Query().Get("v") == ver {
		cache = "public, max-age=31536000, immutable"
	}
	hd := w.Header()
	hd.Set("Cache-Control", cache)
	hd.Set("ETag", etag)
	for _, t := range strings.Split(r.Header.Get("If-None-Match"), ",") {
		if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), "W/")) == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	hd.Set("Content-Type", mime)
	hd.Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func (h *Handler) internalCategories(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListInternalCategories(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]InternalCategoryDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toInternalCategory(c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var in service.CategoryRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := h.svc.CreateCategory(r.Context(), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toInternalCategory(c))
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.CategoryRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := h.svc.UpdateCategory(r.Context(), id, in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toInternalCategory(c))
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) uploadCategoryCover(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	data, ok := readUploadFile(w, r)
	if !ok {
		return
	}
	c, err := h.svc.UploadCategoryCover(r.Context(), id, data)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toInternalCategory(c))
}

func (h *Handler) deleteCategoryCover(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	c, err := h.svc.DeleteCategoryCover(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toInternalCategory(c))
}
