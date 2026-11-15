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

// CartUpdate is the answer to a change: the cart as now stored, plus how many
// units of the changed offer are left for everyone else.
type CartUpdate struct {
	Cart    domain.Order `json:"cart"`
	OfferID string       `json:"offerId"`
	Stock   int          `json:"stock"`
}

// errNotExpired: the line was touched again after we listed it; leave it alone.
var errNotExpired = errors.New("reservation not expired")

// CartStore is what the cart needs: reservations, plus reading the cart order.
type CartStore interface {
	repository.Reservations
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
}

type Cart struct {
	repo CartStore
}

func NewCart(repo CartStore) *Cart { return &Cart{repo: repo} }

// SetQuantity makes the cart hold exactly `qty` units of an offer (0 = remove).
// Stock (and that many keys) is reserved the moment an item goes into the cart
// and returned when it leaves. Setting an absolute number (not "+1") makes
// retries harmless. A paid cart returns repository.ErrCartClosed.
func (c *Cart) SetQuantity(ctx context.Context, cartID, offerID string, qty int) (*CartUpdate, error) {
	if !cartIDPattern.MatchString(cartID) {
		return nil, ErrInvalidCartID
	}
	if qty < 0 || qty > MaxPerLine {
		return nil, ErrInvalidQuantity
	}
	var stockLeft int
	cart, err := c.repo.UpdateReservation(ctx, cartID, offerID, func(stock int, cur domain.Reservation) (int, domain.Reservation, error) {
		newStock, err := reserve(stock, cur.Quantity, qty)
		if err != nil {
			return 0, domain.Reservation{}, err
		}
		stockLeft = newStock
		return newStock, domain.Reservation{Quantity: qty, ExpiresAt: time.Now().Add(ReservationTTL)}, nil
	})
	if err != nil {
		return nil, err
	}
	return &CartUpdate{Cart: cart.Public(), OfferID: offerID, Stock: stockLeft}, nil
}

// Get returns the cart order as stored (an empty cart if it has no lines yet).
// The browser renders the cart from this. A paid cart is ErrCartClosed.
func (c *Cart) Get(ctx context.Context, cartID string) (*domain.Order, error) {
	if !cartIDPattern.MatchString(cartID) {
		return nil, ErrInvalidCartID
	}
	o, err := c.repo.GetOrder(ctx, cartID)
	if errors.Is(err, repository.ErrNotFound) {
		return &domain.Order{ID: cartID, Status: domain.OrderCart, Items: []domain.OrderItem{}}, nil
	}
	if err != nil {
		return nil, err
	}
	if o.Status != domain.OrderCart {
		return nil, repository.ErrCartClosed
	}
	pub := o.Public()
	return &pub, nil
}

// ReleaseExpired puts every cart line whose reservation ended back into stock
// and returns how many lines it released. The API runs it every minute.
// Each release is its own transaction that re-checks the expiry, so a line
// the shopper touched a moment ago (after we listed it) is skipped.
func (c *Cart) ReleaseExpired(ctx context.Context, now time.Time) (int, error) {
	expired, err := c.repo.ListExpiredReservations(ctx, now)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, r := range expired {
		_, err := c.repo.UpdateReservation(ctx, r.CartID, r.OfferID, func(stock int, cur domain.Reservation) (int, domain.Reservation, error) {
			if cur.Quantity == 0 || !cur.ExpiresAt.Before(now) {
				return 0, domain.Reservation{}, errNotExpired
			}
			return stock + cur.Quantity, domain.Reservation{}, nil
		})
		switch {
		case errors.Is(err, errNotExpired), errors.Is(err, repository.ErrNotFound), errors.Is(err, repository.ErrCartClosed):
			// touched again, already gone, offer deleted, or paid meanwhile: skip
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
