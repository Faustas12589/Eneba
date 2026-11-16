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

// Favorites over HTTP as a guest: add (twice), list, remove; unknown product 404; no user 401.
func TestFavoritesRoutes(t *testing.T) {
	r := newTestRouter(t)
	const guest = "0123456789abcdef0123456789abcdef"
	do := func(method, path, guestID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if guestID != "" {
			req.Header.Set("X-Guest-Id", guestID)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	for _, step := range []struct {
		method, path, guestID string
		want                  int
	}{
		{http.MethodGet, "/api/favorites", "", http.StatusUnauthorized},
		{http.MethodPut, "/api/favorites/p001", guest, http.StatusNoContent},
		{http.MethodPut, "/api/favorites/p001", guest, http.StatusNoContent}, // repeat is fine
		{http.MethodPut, "/api/favorites/nope", guest, http.StatusNotFound},
	} {
		if rec := do(step.method, step.path, step.guestID); rec.Code != step.want {
			t.Fatalf("%s %s: got %d, want %d", step.method, step.path, rec.Code, step.want)
		}
	}
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.Unmarshal(do(http.MethodGet, "/api/favorites", guest).Body.Bytes(), &list)
	if len(list.Items) != 1 || list.Items[0].ID != "p001" {
		t.Fatalf("favorites: %+v", list.Items)
	}
	if rec := do(http.MethodDelete, "/api/favorites/p001", guest); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", rec.Code)
	}
	_ = json.Unmarshal(do(http.MethodGet, "/api/favorites", guest).Body.Bytes(), &list)
	if len(list.Items) != 0 {
		t.Fatalf("after delete: %+v", list.Items)
	}
}
