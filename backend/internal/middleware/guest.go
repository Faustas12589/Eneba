package middleware

import (
	"context"
	"net/http"
	"regexp"
)

// GuestHeader carries a random id the browser makes once and keeps in
// localStorage. It identifies a visitor who has not signed in (DECISIONS #20, #28).
const GuestHeader = "X-Guest-Id"

// Same shape rule as cart ids: long and random, so it can't be guessed.
var guestIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{16,64}$`)

// ValidGuestID reports whether id looks like a browser guest id.
func ValidGuestID(id string) bool { return guestIDPattern.MatchString(id) }

// GuestUserID is the user id a guest's profile and orders are stored under.
func GuestUserID(id string) string { return "guest:" + id }

// RequireGuest makes the request's user "guest:<id>", so handlers use UserFrom
// the same way for guests and signed-in users. Anyone holding the id sees that
// visitor's data, like anyone holding a cart id can change that cart.
// Routes use it through Identify, which prefers a login token.
func RequireGuest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(GuestHeader)
			if !ValidGuestID(id) {
				http.Error(w, `{"error":"missing or invalid X-Guest-Id header"}`, http.StatusUnauthorized)
				return
			}
			u := &User{UID: GuestUserID(id)}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
		})
	}
}
