package router

import (
	"encoding/json"
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

// The homepage is whatever GET /api/home lists: every block in it must have a working
// endpoint that answers its own id, its items, and a title if it is a product rail.
func TestHomeLayoutBlocks(t *testing.T) {
	r := newTestRouter(t)
	get := func(path string, out any) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Guest-Id", "0123456789abcdef0123456789abcdef") // personal blocks need a user
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: %d %s", path, rec.Code, rec.Body)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
	}

	var layout struct {
		Blocks []struct{ ID, Type string } `json:"blocks"`
	}
	get("/api/home", &layout)
	if len(layout.Blocks) == 0 {
		t.Fatal("empty homepage layout")
	}
	for _, b := range layout.Blocks {
		var block struct {
			ID    string            `json:"id"`
			Title string            `json:"title"`
			Items []json.RawMessage `json:"items"`
		}
		get("/api/home/"+b.ID, &block)
		if block.ID != b.ID || block.Items == nil {
			t.Errorf("block %s: got id %q, items %v", b.ID, block.ID, block.Items)
		}
		if b.Type == "products" && block.Title == "" {
			t.Errorf("product block %s has no title", b.ID)
		}
	}
}
