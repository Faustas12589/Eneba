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
