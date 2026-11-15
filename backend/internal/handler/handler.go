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
	Catalog *service.Catalog
	Cart    *service.Cart
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

func (h *Handler) Banners(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.Banners(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) BestSellers(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.BestSellers(r.Context(), 8)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) Deals(w http.ResponseWriter, r *http.Request) {
	items, err := h.Catalog.Deals(r.Context(), 8)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// Personal blocks: need the request's user (guest id while login is off).

// RecentlyViewed: GET /api/home/recently-viewed -> {"items": [product...]}, newest first.
func (h *Handler) RecentlyViewed(w http.ResponseWriter, r *http.Request) {
	items, err := h.Profiles.RecentlyViewed(r.Context(), middleware.UserFrom(r.Context()).UID, 8)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// Recommended: GET /api/home/recommended -> {"items": [product...]}.
func (h *Handler) Recommended(w http.ResponseWriter, r *http.Request) {
	items, err := h.Profiles.Recommended(r.Context(), middleware.UserFrom(r.Context()).UID, 8)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
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

// Me: GET /api/me (requires Firebase ID token)
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFrom(r.Context())
	writeJSON(w, http.StatusOK, map[string]string{"uid": u.UID, "email": u.Email})
}

// Checkout: POST /api/orders  body {"cartId": "..."} -> 201 with the delivered order.
// Simulated payment: the cart's reserved keys are sold to the request's user
// (for now a browser guest id, see router).
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
