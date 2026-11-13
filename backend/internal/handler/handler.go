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
		Q:         q.Get("q"),
		Platforms: csv(q.Get("platform")),
		Regions:   csv(q.Get("region")),
		Genres:    csv(q.Get("genre")),
		MinPrice:  atoi(q.Get("minPrice")),
		MaxPrice:  atoi(q.Get("maxPrice")),
		Sort:      q.Get("sort"),
		Page:      atoi(q.Get("page")),
		PageSize:  atoi(q.Get("pageSize")),
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

// --- Auth-protected ---------------------------------------------------

// Me: GET /api/me (requires Firebase ID token)
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFrom(r.Context())
	writeJSON(w, http.StatusOK, map[string]string{"uid": u.UID, "email": u.Email})
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
