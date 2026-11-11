package handler

import (
	"net/http"

	mw "donghai/internal/middleware"
)

// StaffStatDTO là số liệu nhập liệu của 1 tài khoản (chỉ admin thấy).
type StaffStatDTO struct {
	UserID           int64  `json:"user_id"`
	Username         string `json:"username"`
	FullName         string `json:"full_name"`
	Role             string `json:"role"`
	IsActive         bool   `json:"is_active"`
	CustomersTotal   int    `json:"customers_total"`
	OrdersTotal      int    `json:"orders_total"`
	OrdersThisMonth  int    `json:"orders_this_month"`
	RevenueThisMonth int64  `json:"revenue_this_month"`
}

// DashboardDTO là số liệu tổng quan theo phạm vi người xem.
type DashboardDTO struct {
	Scope            string         `json:"scope"` // "all" | "own"
	CustomersTotal   int            `json:"customers_total"`
	OrdersThisMonth  int            `json:"orders_this_month"`
	OrdersTotal      int            `json:"orders_total"`
	RevenueThisMonth int64          `json:"revenue_this_month"`
	RecentOrders     []OrderDTO     `json:"recent_orders"`
	StaffStats       []StaffStatDTO `json:"staff_stats,omitempty"`
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.GetDashboard(r.Context(), mw.UserFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	today := h.svc.Today()
	out := DashboardDTO{Scope: d.Scope, CustomersTotal: d.CustomersTotal, OrdersThisMonth: d.OrdersThisMonth,
		OrdersTotal: d.OrdersTotal, RevenueThisMonth: d.RevenueThisMonth, RecentOrders: []OrderDTO{}}
	for _, o := range d.RecentOrders {
		out.RecentOrders = append(out.RecentOrders, toOrder(o, today, false))
	}
	for _, x := range d.StaffStats {
		out.StaffStats = append(out.StaffStats, StaffStatDTO{UserID: x.UserID, Username: x.Username, FullName: x.FullName,
			Role: x.Role, IsActive: x.Active, CustomersTotal: x.CustomersTotal, OrdersTotal: x.OrdersTotal,
			OrdersThisMonth: x.OrdersInRange, RevenueThisMonth: x.RevenueInRange})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) internalProducts(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListInternalProducts(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]InternalProductDTO, 0, len(list))
	for _, p := range list {
		out = append(out, toInternalProduct(p, false))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) internalProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.svc.GetInternalProduct(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toInternalProduct(p, true))
}

func (h *Handler) priceList(w http.ResponseWriter, r *http.Request) {
	pl, err := h.svc.GetPriceList(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, toPriceList(pl))
}

func (h *Handler) getPromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	p, err := h.svc.GetPromotion(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPromotion(p, h.svc.Today()))
}
