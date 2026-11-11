package handler

import (
	"net/http"
	"time"

	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

// LoginDTO là kết quả đăng nhập.
type LoginDTO struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserDTO   `json:"user"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in service.LoginRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	s, err := h.svc.Login(r.Context(), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, LoginDTO{Token: s.Token, ExpiresAt: s.ExpiresAt, User: toUser(s.User)})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, toUser(mw.UserFromContext(r.Context())))
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var in service.ChangePasswordRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	s, err := h.svc.ChangePassword(r.Context(), mw.UserFromContext(r.Context()), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"token": s.Token, "expires_at": s.ExpiresAt})
}
