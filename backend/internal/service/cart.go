package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

// MaxPerLine caps one cart line, so one cart can't hoard an offer's whole stock.
const MaxPerLine = 10

var (
	ErrInvalidCartID   = errors.New("invalid cart id")
	ErrInvalidQuantity = fmt.Errorf("quantity must be between 0 and %d", MaxPerLine)
)

// Cart ids are random UUIDs made by the browser. Anyone who knows one can
// change that cart, so they must be long and unguessable; we only check shape.
var cartIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{16,64}$`)

// OutOfStockError: the cart asked for more than is available. Available is how
// many more units exist right now; Reserved is what this cart still holds.
type OutOfStockError struct {
	Available int
	Reserved  int
}

func (e *OutOfStockError) Error() string {
	return fmt.Sprintf("not enough stock: %d more available", e.Available)
}

// ReservationTTL: a cart line is held this long after it was last changed.
// Then the sweeper (ReleaseExpired) puts the units back in stock, so abandoned
// carts can't block an offer forever.
const ReservationTTL = 30 * time.Minute

// CartLine is the answer to a change: what the cart holds for one offer and
// how many units are left for everyone else.
type CartLine struct {
	OfferID   string    `json:"offerId"`
	Quantity  int       `json:"quantity"`
	Stock     int       `json:"stock"`
	ExpiresAt time.Time `json:"expiresAt,omitzero"` // zero when the line was removed
}

// ReservedLine is one line of GET /api/carts/{id}: what the server still holds.
type ReservedLine struct {
	OfferID   string    `json:"offerId"`
	Quantity  int       `json:"quantity"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// errNotExpired: the line was touched again after we listed it; leave it alone.
var errNotExpired = errors.New("reservation not expired")

type Cart struct {
	repo repository.Reservations
}

func NewCart(repo repository.Reservations) *Cart { return &Cart{repo: repo} }

// SetQuantity makes the cart hold exactly `qty` units of an offer (0 = remove).
// Stock is reserved the moment an item goes into the cart and returned when it
// leaves. Setting an absolute number (not "+1") makes retries harmless.
func (c *Cart) SetQuantity(ctx context.Context, cartID, offerID string, qty int) (*CartLine, error) {
	if !cartIDPattern.MatchString(cartID) {
		return nil, ErrInvalidCartID
	}
	if qty < 0 || qty > MaxPerLine {
		return nil, ErrInvalidQuantity
	}
	var line CartLine
	err := c.repo.UpdateReservation(ctx, cartID, offerID, func(stock int, cur domain.Reservation) (int, int, error) {
		newStock, err := reserve(stock, cur.Quantity, qty)
		if err != nil {
			return 0, 0, err
		}
		line = CartLine{OfferID: offerID, Quantity: qty, Stock: newStock}
		if qty > 0 {
			line.ExpiresAt = time.Now().Add(ReservationTTL)
		}
		return newStock, qty, nil
	})
	if err != nil {
		return nil, err
	}
	return &line, nil
}

// Lines returns what the server still holds for a cart, so the browser can
// drop lines whose reservation expired.
func (c *Cart) Lines(ctx context.Context, cartID string) ([]ReservedLine, error) {
	if !cartIDPattern.MatchString(cartID) {
		return nil, ErrInvalidCartID
	}
	rs, err := c.repo.ListCartReservations(ctx, cartID)
	if err != nil {
		return nil, err
	}
	lines := make([]ReservedLine, 0, len(rs))
	for _, r := range rs {
		lines = append(lines, ReservedLine{OfferID: r.OfferID, Quantity: r.Quantity, ExpiresAt: r.UpdatedAt.Add(ReservationTTL)})
	}
	return lines, nil
}

// ReleaseExpired puts every reservation older than ReservationTTL back into
// stock and returns how many lines it released. The API runs it every minute.
// Each release is its own transaction that re-checks the timestamp, so a line
// the shopper touched a moment ago (after we listed it) is skipped.
func (c *Cart) ReleaseExpired(ctx context.Context, now time.Time) (int, error) {
	cutoff := now.Add(-ReservationTTL)
	expired, err := c.repo.ListReservationsBefore(ctx, cutoff)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, r := range expired {
		err := c.repo.UpdateReservation(ctx, r.CartID, r.OfferID, func(stock int, cur domain.Reservation) (int, int, error) {
			if cur.Quantity == 0 || cur.UpdatedAt.After(cutoff) {
				return 0, 0, errNotExpired
			}
			return stock + cur.Quantity, 0, nil
		})
		switch {
		case errors.Is(err, errNotExpired), errors.Is(err, repository.ErrNotFound):
			// touched again, already gone, or the offer was deleted: skip
		case err != nil:
			return released, err
		default:
			released++
		}
	}
	return released, nil
}

// reserve moves units between the offer's stock and the cart. Going from
// `reserved` to `want` needs (want - reserved) more units; a negative number
// gives units back. Returns the offer's new stock.
func reserve(stock, reserved, want int) (int, error) {
	need := want - reserved
	if need > stock {
		return 0, &OutOfStockError{Available: stock, Reserved: reserved}
	}
	return stock - need, nil
}
