package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eneba-marketplace/backend/internal/handler"
	"eneba-marketplace/backend/internal/middleware"
	"eneba-marketplace/backend/internal/repository"
	"eneba-marketplace/backend/internal/service"
)

// newTestRouter wires the real router to the memory store and the dev verifier,
// which accepts only the bearer token "dev" (DECISIONS #10).
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	repo, err := repository.NewMemory()
	if err != nil {
		t.Fatal(err)
	}
	catalog := service.NewCatalog(repo)
	h := &handler.Handler{
		Catalog:  catalog,
		Cart:     service.NewCart(repo),
		Orders:   service.NewOrders(repo),
		Profiles: service.NewProfiles(repo, repo, catalog),
	}
	return New(h, middleware.DevVerifier{}, nil)
}

// The assignment: checkout and order history are only for signed-in users.
// A guest id is not enough for them; it is enough for the personal homepage blocks.
func TestSignedInOnlyRoutes(t *testing.T) {
	r := newTestRouter(t)
	const guest = "0123456789abcdef0123456789abcdef"

	cases := []struct {
		name, method, path, token, guestID string
		want                               int
	}{
		{"orders: nobody", http.MethodGet, "/api/orders", "", "", http.StatusUnauthorized},
		{"orders: guest id", http.MethodGet, "/api/orders", "", guest, http.StatusUnauthorized},
		{"orders: bad token", http.MethodGet, "/api/orders", "nope", "", http.StatusUnauthorized},
		{"orders: signed in", http.MethodGet, "/api/orders", "dev", "", http.StatusOK},
		{"checkout: guest id", http.MethodPost, "/api/orders", "", guest, http.StatusUnauthorized},
		{"me: guest id", http.MethodGet, "/api/me", "", guest, http.StatusUnauthorized},
		{"recommended: guest id", http.MethodGet, "/api/home/recommended", "", guest, http.StatusOK},
		{"products: nobody", http.MethodGet, "/api/products", "", "", http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.path, nil)
			if c.token != "" {
				req.Header.Set("Authorization", "Bearer "+c.token)
			}
			if c.guestID != "" {
				req.Header.Set("X-Guest-Id", c.guestID)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("%s %s: got %d, want %d (body %s)", c.method, c.path, rec.Code, c.want, rec.Body)
			}
		})
	}
}
