package handler

import (
	"net/http"
)

// DashboardDTO là số liệu tổng quan.
type DashboardDTO struct {
	CustomersTotal  int        `json:"customers_total"`
	OrdersThisMonth int        `json:"orders_this_month"`
	OrdersTotal     int        `json:"orders_total"`
	RecentOrders    []OrderDTO `json:"recent_orders"`
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	today := h.svc.Today()
	out := DashboardDTO{CustomersTotal: d.CustomersTotal, OrdersThisMonth: d.OrdersThisMonth,
		OrdersTotal: d.OrdersTotal, RecentOrders: []OrderDTO{}}
	for _, o := range d.RecentOrders {
		out.RecentOrders = append(out.RecentOrders, toOrder(o, today, false))
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
