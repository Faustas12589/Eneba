package middleware

import (
	"context"
	"net/http"
	"regexp"
)

// GuestHeader carries a random id the browser makes once and keeps in
// localStorage. It stands in for a signed-in user while login is switched off.
const GuestHeader = "X-Guest-Id"

// Same shape rule as cart ids: long and random, so it can't be guessed.
var guestIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{16,64}$`)

// RequireGuest is the temporary replacement for RequireAuth (DECISIONS #20).
// The request's user is "guest:<id>", so handlers keep using UserFrom and
// orders are still per buyer. Anyone holding the id sees that buyer's orders,
// like anyone holding a cart id can change that cart.
// To bring login back, use RequireAuth(verifier) in the router again.
func RequireGuest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(GuestHeader)
			if !guestIDPattern.MatchString(id) {
				http.Error(w, `{"error":"missing or invalid X-Guest-Id header"}`, http.StatusUnauthorized)
				return
			}
			u := &User{UID: "guest:" + id}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
		})
	}
}
