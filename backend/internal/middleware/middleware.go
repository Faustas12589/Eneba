// Package middleware holds HTTP middleware: CORS and Firebase auth.
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
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
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
			tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if tok == "" || tok == r.Header.Get("Authorization") {
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

func UserFrom(ctx context.Context) *User {
	u, _ := ctx.Value(ctxKey{}).(*User)
	return u
}
