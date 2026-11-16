package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuthWithDevVerifier(t *testing.T) {
	var gotUID string
	h := RequireAuth(DevVerifier{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUID = UserFrom(r.Context()).UID
	}))

	cases := []struct {
		header string
		want   int
	}{
		{"", http.StatusUnauthorized},             // no header
		{"dev", http.StatusUnauthorized},          // missing "Bearer "
		{"Bearer wrong", http.StatusUnauthorized}, // unknown token
		{"Bearer " + DevToken, http.StatusOK},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		if c.header != "" {
			req.Header.Set("Authorization", c.header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("Authorization %q: got %d, want %d", c.header, rec.Code, c.want)
		}
	}
	if gotUID != "dev-user" {
		t.Errorf("handler saw uid %q, want dev-user", gotUID)
	}
}

func TestIdentify(t *testing.T) {
	var gotUID string
	h := Identify(DevVerifier{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUID = UserFrom(r.Context()).UID
	}))
	const guest = "0f8c2a1e-5b7d-4c3a-9e2f-1a2b3c4d5e6f"

	cases := []struct {
		name, auth, guest string
		wantCode          int
		wantUID           string
	}{
		{"token wins over guest id", "Bearer " + DevToken, guest, http.StatusOK, "dev-user"},
		{"guest only", "", guest, http.StatusOK, "guest:" + guest},
		{"bad token is not downgraded to guest", "Bearer wrong", guest, http.StatusUnauthorized, ""},
		{"nothing", "", "", http.StatusUnauthorized, ""},
	}
	for _, c := range cases {
		gotUID = ""
		req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
		if c.auth != "" {
			req.Header.Set("Authorization", c.auth)
		}
		if c.guest != "" {
			req.Header.Set(GuestHeader, c.guest)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.wantCode || gotUID != c.wantUID {
			t.Errorf("%s: got %d %q, want %d %q", c.name, rec.Code, gotUID, c.wantCode, c.wantUID)
		}
	}
}
