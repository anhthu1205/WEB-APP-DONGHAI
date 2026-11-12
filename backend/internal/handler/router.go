// Package handler: HTTP — decode/encode JSON, gọi service. Không chứa SQL.
package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	mw "donghai/internal/middleware"
	"donghai/internal/service"
)

// Limits cấu hình rate limit (nil = tắt, dùng trong test).
type Limits struct {
	General *mw.Limiter
	Login   *mw.Limiter
	Lookup  *mw.Limiter
}

// DefaultLimits theo phương án: lookup 5/phút burst 5 + 300/giờ toàn cục; login 10/phút burst 5;
// toàn bộ /api/* (trừ /api/images/*) 120/phút.
func DefaultLimits() Limits {
	return Limits{
		General: mw.NewLimiter(120, 120),
		Login:   mw.NewLimiter(10, 5),
		Lookup:  mw.NewLimiter(5, 5).WithGlobal(300, 300),
	}
}

// Options cấu hình router.
type Options struct {
	Service        *service.Service
	Logger         *slog.Logger
	StaticDir      string
	AllowedOrigins []string
	Limits         Limits
}

// Handler giữ phụ thuộc cho các route.
type Handler struct {
	svc *service.Service
}

// NewRouter khai báo toàn bộ route.
func NewRouter(o Options) http.Handler {
	h := &Handler{svc: o.Service}
	log := o.Logger
	if log == nil {
		log = slog.Default()
	}

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(mw.Logger(log))
	r.Use(mw.Recoverer(log))
	r.Use(mw.SecureHeaders)
	r.Use(mw.CORS(o.AllowedOrigins))

	r.Route("/api", func(api chi.Router) {
		api.Use(chimw.Timeout(30 * time.Second))

		// Ảnh: ngoài rate limit chung (K6)
		api.Get("/images/{uuid}", h.getImage)
		api.Get("/public/logo", h.getLogo)

		api.Group(func(g chi.Router) {
			g.Use(mw.RateLimit(o.Limits.General))

			g.Get("/health", h.health)

			g.Route("/public", func(p chi.Router) {
				p.Get("/company", h.publicCompany)
				p.Get("/products", h.publicProducts)
				p.Get("/products/{slug}", h.publicProduct)
				p.With(mw.RateLimit(o.Limits.Lookup)).Post("/warranty/lookup", h.warrantyLookup)
			})

			g.Route("/auth", func(a chi.Router) {
				a.With(mw.RateLimit(o.Limits.Login)).Post("/login", h.login)
				a.Group(func(a chi.Router) {
					a.Use(mw.RequireAuth(h.svc))
					a.Get("/me", h.me)
					a.Post("/change-password", h.changePassword)
				})
			})

			g.Route("/internal", func(in chi.Router) {
				in.Use(mw.RequireAuth(h.svc))
				in.Get("/dashboard", h.dashboard)
				in.Get("/products", h.internalProducts)
				in.Get("/products/{id}", h.internalProduct)
				in.Get("/price-list", h.priceList)
				in.Get("/promotions/{id}", h.getPromotion)
				in.Get("/customers", h.listCustomers)
				in.Get("/customers/by-phone", h.customerByPhone)
				in.Get("/customers/{id}", h.getCustomer)
				in.Post("/customers", h.createCustomer)
				in.Get("/orders", h.listOrders)
				in.Get("/orders/{id}", h.getOrder)
				in.Post("/orders", h.createOrder)
			})

			g.Route("/admin", func(ad chi.Router) {
				ad.Use(mw.RequireAuth(h.svc))
				ad.Use(mw.RequireAdmin)
				ad.Post("/products", h.createProduct)
				ad.Put("/products/{id}", h.updateProduct)
				ad.Post("/products/{id}/duplicate", h.duplicateProduct)
				ad.Patch("/products/{id}/visibility", h.setVisibility)
				ad.Delete("/products/{id}", h.deleteProduct)
				ad.Post("/products/{id}/images", h.uploadImage)
				ad.Patch("/images/{uuid}", h.setMainImage)
				ad.Delete("/images/{uuid}", h.deleteImage)
				ad.Put("/company", h.updateCompany)
				ad.Post("/company/logo", h.uploadLogo)
				ad.Delete("/company/logo", h.deleteLogo)
				ad.Get("/promotions", h.listPromotions)
				ad.Post("/promotions", h.createPromotion)
				ad.Put("/promotions/{id}", h.updatePromotion)
				ad.Delete("/promotions/{id}", h.deletePromotion)
				ad.Put("/customers/{id}", h.updateCustomer)
				ad.Delete("/customers/{id}", h.deleteCustomer)
				ad.Put("/orders/{id}", h.updateOrder)
				ad.Delete("/orders/{id}", h.deleteOrder)
				ad.Get("/users", h.listUsers)
				ad.Post("/users", h.createUser)
				ad.Patch("/users/{id}", h.updateUser)
				ad.Post("/users/{id}/reset-password", h.resetPassword)
			})
		})

		// Mọi route lạ dưới /api → 404/405 JSON (không rơi vào SPA)
		api.NotFound(func(w http.ResponseWriter, _ *http.Request) {
			mw.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy API")
		})
		api.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
			mw.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ")
		})
	})

	static := NewStatic(o.StaticDir)
	r.NotFound(static.ServeHTTP)
	r.MethodNotAllowed(static.ServeHTTP)
	return r
}
