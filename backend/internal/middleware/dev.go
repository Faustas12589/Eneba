package middleware

import (
	"context"
	"errors"
)

// DevToken is the bearer token the frontend sends when Firebase is not configured.
const DevToken = "dev"

// DevVerifier lets signed-in routes work offline (STORE=memory, see DECISIONS #10).
// It accepts only DevToken and maps it to one fixed user. main.go wires it ONLY in
// memory mode; production (STORE=firestore) always uses FirebaseVerifier.
type DevVerifier struct{}

func (DevVerifier) Verify(_ context.Context, idToken string) (*User, error) {
	if idToken != DevToken {
		return nil, errors.New("dev verifier: unknown token")
	}
	return &User{UID: "dev-user", Email: "dev@localhost"}, nil
}
