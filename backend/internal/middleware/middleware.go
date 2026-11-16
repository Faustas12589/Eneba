// Package middleware holds HTTP middleware: CORS, Firebase auth and guest ids.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

func CORS(origins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", GuestHeader},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

// User is the identity extracted from a verified Firebase ID token.
type User struct {
	UID   string
	Email string
}

// TokenVerifier is implemented by FirebaseVerifier; tests can fake it.
type TokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*User, error)
}

type ctxKey struct{}

// RequireAuth rejects requests without a valid "Authorization: Bearer <idToken>".
func RequireAuth(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v == nil {
				http.Error(w, `{"error":"auth not configured"}`, http.StatusServiceUnavailable)
				return
			}
			tok := bearer(r)
			if tok == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}
			u, err := v.Verify(r.Context(), tok)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
		})
	}
}

// Identify is for routes guests may use too (order history, profile blocks).
// A request with a bearer token is the signed-in user (a bad token is 401, never
// a silent fall back to guest); without one it is the browser's guest id.
func Identify(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		asUser, asGuest := RequireAuth(v)(next), RequireGuest()(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bearer(r) != "" {
				asUser.ServeHTTP(w, r)
				return
			}
			asGuest.ServeHTTP(w, r)
		})
	}
}

// bearer returns the token from "Authorization: Bearer <token>", or "" if there is none.
func bearer(r *http.Request) string {
	tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return ""
	}
	return tok
}

func UserFrom(ctx context.Context) *User {
	u, _ := ctx.Value(ctxKey{}).(*User)
	return u
}
