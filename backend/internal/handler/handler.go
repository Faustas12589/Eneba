// Package handler is the HTTP layer: parse request, call a service, write JSON.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"eneba-marketplace/backend/internal/middleware"
	"eneba-marketplace/backend/internal/repository"
	"eneba-marketplace/backend/internal/service"
)

type Handler struct {
	Catalog  *service.Catalog
	Cart     *service.Cart
	Orders   *service.Orders
	Profiles *service.Profiles
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	log.Printf("error: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Catalog / search -------------------------------------------------

// ListProducts: GET /api/products?q=&platform=a,b&region=&genre=&minPrice=&maxPrice=&sort=&page=&pageSize=
// minPrice/maxPrice are in cents.
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pg, err := h.Catalog.List(r.Context(), service.Filter{
		Q:          q.Get("q"),
		Platforms:  csv(q.Get("platform")),
		Regions:    csv(q.Get("region")),
		Genres:     csv(q.Get("genre")),
		Categories: csv(q.Get("category")),
		MinPrice:   atoi(q.Get("minPrice")),
		MaxPrice:   atoi(q.Get("maxPrice")),
		Sort:       q.Get("sort"),
		Page:       atoi(q.Get("page")),
		PageSize:   atoi(q.Get("pageSize")),
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pg)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	d, err := h.Catalog.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// Suggest: GET /api/search?q=...  (header dropdown, top results only)
func (h *Handler) Suggest(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.Suggest(r.Context(), r.URL.Query().Get("q"), 6)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// --- Homepage blocks: one endpoint per block -----------------------------
// Every block answers {"id", "title", "info", "moreUrl", "items"}: its text
// (service.HomeBlock) plus its items, so the frontend hardcodes nothing (DECISIONS #35).

// Home: GET /api/home -> {"blocks": [{"id", "type", "personal"}...]}, top to bottom.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"blocks": service.HomeLayout})
}

// Banners: GET /api/home/banners -> the hero tiles.
func (h *Handler) Banners(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.Banners(r.Context())
	writeBlock(w, "banners", items, err)
}

// BestSellers: GET /api/home/best-sellers -> most sold products.
func (h *Handler) BestSellers(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.BestSellers(r.Context(), 8)
	writeBlock(w, "best-sellers", items, err)
}

// GiftCards: GET /api/home/gift-cards -> best-selling eGift cards (Amazon, Spotify...).
func (h *Handler) GiftCards(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.TopInCategory(r.Context(), "gift-cards", 8)
	writeBlock(w, "gift-cards", items, err)
}

// TopUps: GET /api/home/top-ups -> best-selling gaming cards (Steam, PSN, Roblox...).
func (h *Handler) TopUps(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.TopInCategory(r.Context(), "gaming-cards", 8)
	writeBlock(w, "top-ups", items, err)
}

// PreOrders: GET /api/home/pre-orders -> unreleased products, soonest first.
func (h *Handler) PreOrders(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.PreOrders(r.Context(), 8)
	writeBlock(w, "pre-orders", items, err)
}

// Categories: GET /api/home/categories -> items [{"category", "count"}...] in menu order.
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.CategoryCounts(r.Context())
	writeBlock(w, "categories", items, err)
}

// Personal blocks: need the request's user (login token or guest id).

// RecentlyViewed: GET /api/home/recently-viewed -> the user's viewed products, newest first.
func (h *Handler) RecentlyViewed(w http.ResponseWriter, r *http.Request) {
	items, err := h.Profiles.RecentlyViewed(r.Context(), middleware.UserFrom(r.Context()).UID, 8)
	writeBlock(w, "recently-viewed", items, err)
}

// Recommended: GET /api/home/recommended -> products picked from the user's history.
func (h *Handler) Recommended(w http.ResponseWriter, r *http.Request) {
	items, err := h.Profiles.Recommended(r.Context(), middleware.UserFrom(r.Context()).UID, 8)
	writeBlock(w, "recommended", items, err)
}

// writeItems answers {"items": [...]} (never null), or the error.
func writeItems[T any](w http.ResponseWriter, items []T, err error) {
	if err != nil {
		writeErr(w, err)
		return
	}
	if items == nil {
		items = []T{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// writeBlock answers a block's text and its items (never null), or the error.
// The embedded BlockText's fields sit next to "items" in the JSON.
func writeBlock[T any](w http.ResponseWriter, id string, items []T, err error) {
	if err != nil {
		writeErr(w, err)
		return
	}
	if items == nil {
		items = []T{}
	}
	writeJSON(w, http.StatusOK, struct {
		service.BlockText
		Items []T `json:"items"`
	}{service.HomeBlock(id), items})
}

// --- Profile ---------------------------------------------------------------

// RecordView: POST /api/profile/views {"productId": "..."} -> 204.
// The product page calls it on open; 404 if the product doesn't exist.
func (h *Handler) RecordView(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductID string `json:"productId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil || body.ProductID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `body must be {"productId": "..."}`})
		return
	}
	if err := h.Profiles.RecordView(r.Context(), middleware.UserFrom(r.Context()).UID, body.ProductID); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Favorites (guests too, DECISIONS #37) ------------------------------------

// Favorites: GET /api/favorites -> {"items": [product...]}, newest first.
func (h *Handler) Favorites(w http.ResponseWriter, r *http.Request) {
	items, err := h.Profiles.Favorites(r.Context(), middleware.UserFrom(r.Context()).UID)
	writeItems(w, items, err)
}

// AddFavorite: PUT /api/favorites/{productId} -> 204. Safe to repeat; 404 unknown
// product; 409 when the list already holds service.MaxFavorites.
func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	h.setFavorite(w, r, true)
}

// RemoveFavorite: DELETE /api/favorites/{productId} -> 204. Safe to repeat.
func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	h.setFavorite(w, r, false)
}

func (h *Handler) setFavorite(w http.ResponseWriter, r *http.Request, on bool) {
	err := h.Profiles.SetFavorite(r.Context(), middleware.UserFrom(r.Context()).UID, chi.URLParam(r, "productId"), on)
	if errors.Is(err, service.ErrFavoritesFull) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Cart (guests too) ---------------------------------------------------

// SetCartItem: PUT /api/carts/{cartId}/items/{offerId}  body {"quantity": n}
// Reserves or releases stock so the cart holds exactly n units (0 = remove).
// 200 -> {"cart": <cart order as stored>, "offerId", "stock": units left}.
// 409 when there isn't enough stock; the body says how many are available.
func (h *Handler) SetCartItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Quantity *int `json:"quantity"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil || body.Quantity == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `body must be {"quantity": n}`})
		return
	}
	line, err := h.Cart.SetQuantity(r.Context(), chi.URLParam(r, "cartId"), chi.URLParam(r, "offerId"), *body.Quantity)

	var oos *service.OutOfStockError
	switch {
	case errors.As(err, &oos):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "not enough stock", "stock": oos.Available, "quantity": oos.Reserved,
		})
	case errors.Is(err, service.ErrInvalidCartID), errors.Is(err, service.ErrInvalidQuantity):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, repository.ErrCartClosed):
		// 410 Gone: this cart is now an order; the browser starts a new cart.
		writeJSON(w, http.StatusGone, map[string]string{"error": err.Error()})
	case err != nil:
		writeErr(w, err)
	default:
		writeJSON(w, http.StatusOK, line)
	}
}

// GetCart: GET /api/carts/{cartId} -> the cart order as stored (status "cart").
// The browser renders the cart from this; expired lines are simply gone.
// 410 once the cart was paid.
func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.Cart.Get(r.Context(), chi.URLParam(r, "cartId"))
	switch {
	case errors.Is(err, service.ErrInvalidCartID):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, repository.ErrCartClosed):
		writeJSON(w, http.StatusGone, map[string]string{"error": err.Error()})
	case err != nil:
		writeErr(w, err)
	default:
		writeJSON(w, http.StatusOK, cart)
	}
}

// --- Auth-protected ---------------------------------------------------

// Me: GET /api/me (signed in) -> the account page's purchase numbers.
// Name, photo and sign-up date come from Firebase Auth in the browser.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFrom(r.Context())
	sum, err := h.Orders.Summary(r.Context(), u.UID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"uid": u.UID, "email": u.Email, "orderCount": sum.OrderCount, "totalSpent": sum.TotalSpent,
	})
}

// MergeGuest: POST /api/me/merge-guest  body {"guestId": "..."} (signed in).
// Called right after sign-in: the browser's guest views and orders move to the account.
func (h *Handler) MergeGuest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GuestID string `json:"guestId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !middleware.ValidGuestID(body.GuestID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `body must be {"guestId": "..."}`})
		return
	}
	u := middleware.UserFrom(r.Context())
	moved, err := h.Profiles.MergeGuest(r.Context(), u.UID, middleware.GuestUserID(body.GuestID))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"movedOrders": moved})
}

// Checkout: POST /api/orders  body {"cartId": "..."} -> 201 with the delivered order.
// Simulated payment: the cart's reserved keys are sold to the signed-in user.
// Safe to retry: a paid-but-undelivered order is finished, a delivered one returned.
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CartID string `json:"cartId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `body must be {"cartId": "..."}`})
		return
	}
	order, err := h.Orders.Checkout(r.Context(), middleware.UserFrom(r.Context()).UID, body.CartID)
	switch {
	case errors.Is(err, service.ErrInvalidCartID), errors.Is(err, service.ErrEmptyCart):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotYourOrder):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}) // don't confirm it exists
	case err != nil:
		writeErr(w, err)
	default:
		writeJSON(w, http.StatusCreated, order)
	}
}

// ListOrders: GET /api/orders?page=&pageSize= -> {"items", "total", "page", "pageSize"}
// for the request's user, newest first.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pg, err := h.Orders.List(r.Context(), middleware.UserFrom(r.Context()).UID, atoi(q.Get("page")), atoi(q.Get("pageSize")))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pg)
}

// --- helpers ----------------------------------------------------------

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func csv(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
