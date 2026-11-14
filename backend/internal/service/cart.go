package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

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

// CartLine is the result of a change: what the cart now holds and how many
// units of the offer are left for everyone else.
type CartLine struct {
	OfferID  string `json:"offerId"`
	Quantity int    `json:"quantity"`
	Stock    int    `json:"stock"`
}

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
	err := c.repo.UpdateReservation(ctx, cartID, offerID, func(stock, reserved int) (int, int, error) {
		newStock, err := reserve(stock, reserved, qty)
		if err != nil {
			return 0, 0, err
		}
		line = CartLine{OfferID: offerID, Quantity: qty, Stock: newStock}
		return newStock, qty, nil
	})
	if err != nil {
		return nil, err
	}
	return &line, nil
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
