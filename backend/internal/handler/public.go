package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.svc.Ping(ctx); err != nil {
		mw.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) publicCompany(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.GetCompany(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	writeJSON(w, http.StatusOK, toPublicCompany(c))
}

func (h *Handler) publicProducts(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListPublicProducts(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]PublicProductDTO, 0, len(list))
	for _, p := range list {
		out = append(out, toPublicProduct(p))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) publicProduct(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetPublicProduct(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPublicProductDetail(p))
}

func (h *Handler) warrantyLookup(w http.ResponseWriter, r *http.Request) {
	var in service.WarrantyLookupRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	res, err := h.svc.LookupWarranty(r.Context(), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, toWarrantyLookup(res))
}
