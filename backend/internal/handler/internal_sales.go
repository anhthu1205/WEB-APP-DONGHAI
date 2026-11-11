package handler

import (
	"net/http"

	"donghai/internal/domain"
	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

// CustomerDetailDTO là khách kèm đơn hàng.
type CustomerDetailDTO struct {
	CustomerDTO
	Orders []OrderDTO `json:"orders"`
}

func (h *Handler) listCustomers(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.ListCustomers(r.Context(), r.URL.Query().Get("q"), pageParams(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPage(page, toCustomer))
}

func (h *Handler) customerByPhone(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.GetCustomerByPhone(r.Context(), r.URL.Query().Get("phone"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toCustomer(c))
}

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	d, err := h.svc.GetCustomerDetail(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	today := h.svc.Today()
	out := CustomerDetailDTO{CustomerDTO: toCustomer(d.Customer), Orders: []OrderDTO{}}
	for _, o := range d.Orders {
		out.Orders = append(out.Orders, toOrder(o, today, false))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var in service.CustomerRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := h.svc.CreateCustomer(r.Context(), mw.UserFromContext(r.Context()), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCustomer(c))
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := h.svc.ListOrders(r.Context(), service.OrderFilterRequest{Q: q.Get("q"), From: q.Get("from"), To: q.Get("to")}, pageParams(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	today := h.svc.Today()
	writeJSON(w, http.StatusOK, toPage(page, func(o *domain.Order) OrderDTO { return toOrder(o, today, false) }))
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r, "id")
	if !ok {
		return
	}
	o, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toOrder(o, h.svc.Today(), true))
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in service.CreateOrderRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	o, reused, err := h.svc.CreateOrder(r.Context(), mw.UserFromContext(r.Context()), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := toOrder(o, h.svc.Today(), true)
	out.CustomerReused = &reused
	writeJSON(w, http.StatusCreated, out)
}
