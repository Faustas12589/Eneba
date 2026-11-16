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
		r.Get("/home", h.Home)
		r.Get("/home/banners", h.Banners)
		r.Get("/home/best-sellers", h.BestSellers)
		r.Get("/home/gift-cards", h.GiftCards)
		r.Get("/home/top-ups", h.TopUps)
		r.Get("/home/pre-orders", h.PreOrders)
		r.Get("/home/categories", h.Categories)
		r.Get("/carts/{cartId}", h.GetCart)
		r.Put("/carts/{cartId}/items/{offerId}", h.SetCartItem)

		// guests and signed-in users: a login token if sent, else the X-Guest-Id
		r.Group(func(r chi.Router) {
			r.Use(middleware.Identify(verifier))
			r.Post("/profile/views", h.RecordView)
			r.Get("/favorites", h.Favorites)
			r.Put("/favorites/{productId}", h.AddFavorite)
			r.Delete("/favorites/{productId}", h.RemoveFavorite)
			r.Get("/home/recently-viewed", h.RecentlyViewed)
			r.Get("/home/recommended", h.Recommended)
		})

		// signed-in only (the assignment: checkout and order history need an account):
		// paying, order history, the account page, and merging guest data on sign-in
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(verifier))
			r.Post("/orders", h.Checkout)
			r.Get("/orders", h.ListOrders)
			r.Get("/me", h.Me)
			r.Post("/me/merge-guest", h.MergeGuest)
		})
	})
	return r
}
