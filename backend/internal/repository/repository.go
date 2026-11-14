// Package repository is the only layer that knows where data lives.
// Services depend on the Catalog interface, never on Firestore directly.
package repository

import (
	"context"
	"errors"
	"time"

	"eneba-marketplace/backend/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Catalog interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListOffers(ctx context.Context, productID string) ([]domain.Offer, error)
	ListBanners(ctx context.Context) ([]domain.Banner, error)
}

// ApplyFunc gets an offer's current stock and the cart's current reservation
// (zero value if none), and returns the new values to save (or an error to
// save nothing). The rule lives in the service; the repository only makes it atomic.
type ApplyFunc func(stock int, current domain.Reservation) (newStock, newReserved int, err error)

type Reservations interface {
	// UpdateReservation reads the offer and the cart's reservation, calls apply,
	// and saves both results in one atomic step (Firestore transaction / mutex).
	// Saving sets the reservation's UpdatedAt to now.
	// apply may run more than once if Firestore retries the transaction.
	// Returns ErrNotFound if the offer doesn't exist.
	UpdateReservation(ctx context.Context, cartID, offerID string, apply ApplyFunc) error
	// ListCartReservations returns everything one cart currently holds.
	ListCartReservations(ctx context.Context, cartID string) ([]domain.Reservation, error)
	// ListReservationsBefore returns reservations last updated before t (expired).
	ListReservationsBefore(ctx context.Context, t time.Time) ([]domain.Reservation, error)
}

// Store is everything the API needs; both Memory and Firestore implement it.
type Store interface {
	Catalog
	Reservations
}
