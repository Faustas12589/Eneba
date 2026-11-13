package middleware

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

type FirebaseVerifier struct{ Client *auth.Client }

func (f FirebaseVerifier) Verify(ctx context.Context, idToken string) (*User, error) {
	t, err := f.Client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	email, _ := t.Claims["email"].(string)
	return &User{UID: t.UID, Email: email}, nil
}
