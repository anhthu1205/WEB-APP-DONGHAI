package handler

import (
	"net/http"

	"donghai/internal/domain"
	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

// ----- Sản phẩm -----

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var in service.ProductRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	p, err := h.svc.CreateProduct(r.Context(), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toInternalProduct(p, true))
}

// duplicateProduct sao chép SP (kèm ảnh); bản sao ở trạng thái ẩn.
func (h *Handler) duplicateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.svc.DuplicateProduct(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toInternalProduct(p, true))
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.ProductRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	p, err := h.svc.UpdateProduct(r.Context(), id, in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toInternalProduct(p, true))
}

func (h *Handler) setVisibility(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.VisibilityRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.svc.SetProductVisibility(r.Context(), id, in); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "is_visible": *in.IsVisible})
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteProduct(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ----- Công ty -----

func (h *Handler) updateCompany(w http.ResponseWriter, r *http.Request) {
	var in service.CompanyRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := h.svc.UpdateCompany(r.Context(), mw.UserFromContext(r.Context()), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPublicCompany(c))
}

// ----- Khuyến mãi -----

func (h *Handler) listPromotions(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.ListPromotions(r.Context(), r.URL.Query().Get("status"), pageParams(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	today := h.svc.Today()
	writeJSON(w, http.StatusOK, toPage(page, func(p *domain.Promotion) PromotionDTO { return toPromotion(p, today) }))
}

func (h *Handler) createPromotion(w http.ResponseWriter, r *http.Request) {
	var in service.PromotionRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	p, err := h.svc.CreatePromotion(r.Context(), mw.UserFromContext(r.Context()), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toPromotion(p, h.svc.Today()))
}

func (h *Handler) updatePromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.PromotionRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	p, err := h.svc.UpdatePromotion(r.Context(), id, in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPromotion(p, h.svc.Today()))
}

func (h *Handler) deletePromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeletePromotion(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ----- Khách hàng & đơn hàng (sửa/xóa) -----

func (h *Handler) updateCustomer(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.CustomerRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := h.svc.UpdateCustomer(r.Context(), id, in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toCustomer(c))
}

func (h *Handler) deleteCustomer(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteCustomer(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.UpdateOrderRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	o, err := h.svc.UpdateOrder(r.Context(), id, in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toOrder(o, h.svc.Today(), true))
}

func (h *Handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteOrder(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ----- Tài khoản -----

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListUsers(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]UserDTO, 0, len(list))
	for _, u := range list {
		out = append(out, toUser(u))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var in service.CreateUserRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	u, err := h.svc.CreateUser(r.Context(), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toUser(u))
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.UpdateUserRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	u, err := h.svc.UpdateUser(r.Context(), mw.UserFromContext(r.Context()), id, in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toUser(u))
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	var in service.ResetPasswordRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.svc.ResetPassword(r.Context(), id, in); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
