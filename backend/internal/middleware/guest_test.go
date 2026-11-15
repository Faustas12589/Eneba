package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireGuest(t *testing.T) {
	var gotUID string
	h := RequireGuest()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUID = UserFrom(r.Context()).UID
	}))

	const id = "0f8c2a1e-5b7d-4c3a-9e2f-1a2b3c4d5e6f"
	cases := []struct {
		header string
		want   int
	}{
		{"", http.StatusUnauthorized},          // no header
		{"short", http.StatusUnauthorized},     // too short to be random
		{"bad id with spaces!!", http.StatusUnauthorized},
		{id, http.StatusOK},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
		if c.header != "" {
			req.Header.Set(GuestHeader, c.header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s %q: got %d, want %d", GuestHeader, c.header, rec.Code, c.want)
		}
	}
	if gotUID != "guest:"+id {
		t.Errorf("handler saw uid %q", gotUID)
	}
}
