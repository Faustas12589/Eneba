// Package repository is the only layer that knows where data lives.
// Services depend on these interfaces, never on Firestore directly.
package repository

import (
	"context"
	"errors"
	"slices"
	"time"

	"eneba-marketplace/backend/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
	// ErrCartClosed: the cart was already paid, so it can't change any more.
	ErrCartClosed = errors.New("cart already checked out")
	// errKeyPoolMismatch means offers.stock and the available keys disagree.
	// Every write changes both together, so this only happens with bad seed data.
	errKeyPoolMismatch = errors.New("key pool does not match offer stock")
)

type Catalog interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListOffers(ctx context.Context, productID string) ([]domain.Offer, error)
	ListBanners(ctx context.Context) ([]domain.Banner, error)
}

// ApplyFunc gets an offer's current stock and the cart's current line as a
// reservation (zero value if none), and returns the new stock and line to save
// (or an error to save nothing). The rule lives in the service; the repository
// only makes it atomic.
type ApplyFunc func(stock int, current domain.Reservation) (newStock int, next domain.Reservation, err error)

// Reservations are the lines of orders that are still carts.
type Reservations interface {
	// UpdateReservation reads the offer and the cart's line, calls apply, and in
	// one atomic step (Firestore transaction / mutex) saves the new stock and
	// line and moves keys between "available" and "reserved" to match.
	// A new line is written whole: product name, image, seller, price and keys.
	// A cart with no lines left is deleted. Returns the cart as saved.
	// apply may run more than once if Firestore retries the transaction.
	// Returns ErrNotFound if the offer doesn't exist, ErrCartClosed if the cart was paid.
	UpdateReservation(ctx context.Context, cartID, offerID string, apply ApplyFunc) (*domain.Order, error)
	// ListExpiredReservations returns cart lines whose reservation ended before now.
	ListExpiredReservations(ctx context.Context, now time.Time) ([]domain.Reservation, error)
}

// PayFunc and DeliverFunc are service rules that change an order's stage.
// The repository reads the inputs, calls the func and saves the results in
// one transaction. DeliverFunc gets the order's keys (by the ids on its lines).
type (
	PayFunc     func(o *domain.Order) error
	DeliverFunc func(o *domain.Order, keys []domain.Key) error
)

type Orders interface {
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	PayOrder(ctx context.Context, id string, apply PayFunc) (*domain.Order, error)
	DeliverOrder(ctx context.Context, id string, apply DeliverFunc) (*domain.Order, error)
	ListUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
	// MoveUserOrders gives every order of user `from` to user `to` in one atomic
	// step and returns how many moved. Used when a guest signs in.
	MoveUserOrders(ctx context.Context, from, to string) (int, error)
}

// ProfileFunc is a service rule that changes a profile (e.g. adds a view).
type ProfileFunc func(p *domain.Profile) error

type Profiles interface {
	// GetProfile returns ErrNotFound for a user who has no profile yet.
	GetProfile(ctx context.Context, userID string) (*domain.Profile, error)
	// UpdateProfile reads the profile (an empty one if missing), calls apply and
	// saves the result in one atomic step. apply may run more than once if
	// Firestore retries the transaction.
	UpdateProfile(ctx context.Context, userID string, apply ProfileFunc) (*domain.Profile, error)
}

// Store is everything the API needs; both Memory and Firestore implement it.
type Store interface {
	Catalog
	Reservations
	Orders
	Profiles
}

// reservationOf is a cart line seen as a reservation (zero value if no line).
func reservationOf(cartID string, it *domain.OrderItem) domain.Reservation {
	r := domain.Reservation{CartID: cartID}
	if it != nil {
		r.OfferID, r.Quantity = it.OfferID, it.Quantity
		if it.ExpiresAt != nil {
			r.ExpiresAt = *it.ExpiresAt
		}
	}
	return r
}

// nextItem is a cart line after a change. A new line (cur == nil) is written
// whole from the offer and product; an existing line keeps its snapshot and
// only changes quantity, expiry and keys. `added` are keys just reserved; if
// the quantity went down, the line's last keys are cut off and returned so
// the caller can make them available again.
func nextItem(cur *domain.OrderItem, offer domain.Offer, product domain.Product, next domain.Reservation, added []domain.Key) (domain.OrderItem, []domain.ItemKey) {
	var it domain.OrderItem
	if cur != nil {
		it = *cur
	} else {
		it = domain.OrderItem{
			OfferID: offer.ID, ProductID: offer.ProductID,
			ProductName: product.Name, ImageURL: product.ImageURL,
			SellerName: offer.SellerName, Price: offer.Price,
		}
	}
	keys := slices.Clone(it.Keys)
	for _, k := range added {
		keys = append(keys, domain.ItemKey{ID: k.ID, Code: k.Code})
	}
	var released []domain.ItemKey
	if next.Quantity < len(keys) {
		released = keys[next.Quantity:]
		keys = keys[:next.Quantity]
	}
	it.Keys, it.Quantity, it.ExpiresAt = keys, next.Quantity, nil
	if next.Quantity > 0 {
		exp := next.ExpiresAt
		it.ExpiresAt = &exp
	}
	return it, released
}

// expiredLines returns the cart lines of o whose reservation ended before now.
func expiredLines(o domain.Order, now time.Time) []domain.Reservation {
	var out []domain.Reservation
	if o.Status != domain.OrderCart {
		return nil
	}
	for i := range o.Items {
		if r := reservationOf(o.ID, &o.Items[i]); r.ExpiresAt.Before(now) {
			out = append(out, r)
		}
	}
	return out
}
