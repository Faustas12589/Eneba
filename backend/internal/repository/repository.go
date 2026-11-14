// Package repository is the only layer that knows where data lives.
// Services depend on the Catalog interface, never on Firestore directly.
package repository

import (
	"context"
	"errors"

	"eneba-marketplace/backend/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Catalog interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListOffers(ctx context.Context, productID string) ([]domain.Offer, error)
	ListBanners(ctx context.Context) ([]domain.Banner, error)
}

// ApplyFunc gets an offer's current stock and how many units a cart already
// reserves, and returns the new values to save (or an error to save nothing).
// The rule lives in the service; the repository only makes it atomic.
type ApplyFunc func(stock, reserved int) (newStock, newReserved int, err error)

type Reservations interface {
	// UpdateReservation reads the offer and the cart's reservation, calls apply,
	// and saves both results in one atomic step (Firestore transaction / mutex).
	// apply may run more than once if Firestore retries the transaction.
	// Returns ErrNotFound if the offer doesn't exist.
	UpdateReservation(ctx context.Context, cartID, offerID string, apply ApplyFunc) error
}

// Store is everything the API needs; both Memory and Firestore implement it.
type Store interface {
	Catalog
	Reservations
}
