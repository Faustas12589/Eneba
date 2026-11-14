package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"eneba-marketplace/backend/internal/handler"
	"eneba-marketplace/backend/internal/middleware"
)

func New(h *handler.Handler, verifier middleware.TokenVerifier, corsOrigins []string) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, chimw.Logger, chimw.Recoverer)
	r.Use(middleware.CORS(corsOrigins))

	r.Get("/healthz", h.Health)

	r.Route("/api", func(r chi.Router) {
		// public
		r.Get("/products", h.ListProducts)
		r.Get("/products/{id}", h.GetProduct)
		r.Get("/search", h.Suggest)
		r.Get("/home/banners", h.Banners)
		r.Get("/home/best-sellers", h.BestSellers)
		r.Get("/home/deals", h.Deals)
		r.Get("/carts/{cartId}", h.GetCart)
		r.Put("/carts/{cartId}/items/{offerId}", h.SetCartItem)

		// signed-in only
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(verifier))
			r.Get("/me", h.Me)
			// TODO: checkout, orders
		})
	})
	return r
}
