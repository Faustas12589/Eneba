package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

var (
	ErrEmptyCart = errors.New("cart is empty")
	// ErrNotYourOrder: someone tried to check out another buyer's cart.
	ErrNotYourOrder = errors.New("order belongs to another user")
	// errKeysMissing: a line's keys are not reserved for this order any more.
	// Reserving keys and stock together makes this impossible unless data is broken.
	errKeysMissing = errors.New("order keys are not reserved for it")
)

type Orders struct {
	repo repository.Orders
}

func NewOrders(repo repository.Orders) *Orders { return &Orders{repo: repo} }

// Checkout moves a cart through cart -> paid -> delivered. Payment is
// simulated (DECISIONS #8). Each step is one transaction and only changes
// status fields: names, prices, total and keys were written when the lines
// went into the cart. If delivery fails after payment, calling again finishes it.
func (s *Orders) Checkout(ctx context.Context, userID, cartID string) (*domain.Order, error) {
	if !cartIDPattern.MatchString(cartID) {
		return nil, ErrInvalidCartID
	}
	now := time.Now()
	_, err := s.repo.PayOrder(ctx, cartID, func(o *domain.Order) error { return pay(o, userID, now) })
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrEmptyCart // a cart with no lines is never stored
	}
	if err != nil {
		return nil, err
	}
	o, err := s.repo.DeliverOrder(ctx, cartID, func(o *domain.Order, keys []domain.Key) error {
		return deliver(o, keys, now)
	})
	if err != nil {
		return nil, err
	}
	pub := o.Public()
	return &pub, nil
}

// OrderPage is one page of a user's order history.
type OrderPage struct {
	Items    []domain.Order `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// List returns one page of a user's orders, newest first, as the API shows them.
// Carts are not listed: they have no user until they are paid.
// All of the user's orders are read, then sorted and cut here: a shopper has
// tens of orders, and it needs no Firestore composite index (DECISIONS #22).
func (s *Orders) List(ctx context.Context, userID string, page, pageSize int) (OrderPage, error) {
	orders, err := s.repo.ListUserOrders(ctx, userID)
	if err != nil {
		return OrderPage{}, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	sort.Slice(orders, func(i, j int) bool {
		ti, tj := paidTime(orders[i]), paidTime(orders[j])
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return orders[i].ID < orders[j].ID // same time: fixed order, so pages don't overlap
	})
	start, end := pageBounds(page, pageSize, len(orders))
	items := make([]domain.Order, 0, end-start) // JSON [] instead of null
	for _, o := range orders[start:end] {
		items = append(items, o.Public())
	}
	return OrderPage{Items: items, Total: len(orders), Page: page, PageSize: pageSize}, nil
}

// Summary is what the account page shows about a user's purchases.
type Summary struct {
	OrderCount int `json:"orderCount"`
	TotalSpent int `json:"totalSpent"` // cents
}

// Summary counts the user's orders and adds up what they paid. Carts are not
// counted: they have no user until they are paid.
func (s *Orders) Summary(ctx context.Context, userID string) (Summary, error) {
	orders, err := s.repo.ListUserOrders(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	sum := Summary{OrderCount: len(orders)}
	for _, o := range orders {
		sum.TotalSpent += o.Total
	}
	return sum, nil
}

func paidTime(o domain.Order) time.Time {
	if o.PaidAt == nil {
		return time.Time{}
	}
	return *o.PaidAt
}

// pay is the "cart -> paid" step: it records who bought it and when, and ends
// the reservation timer (a paid line no longer expires). Nothing else changes:
// the price is the one the cart showed when the line was added.
// An order that is already paid is left alone, so a retried checkout continues.
func pay(o *domain.Order, userID string, now time.Time) error {
	if o.Status != domain.OrderCart {
		if o.UserID != userID {
			return ErrNotYourOrder
		}
		return nil
	}
	if len(o.Items) == 0 {
		return ErrEmptyCart
	}
	for i := range o.Items {
		o.Items[i].ExpiresAt = nil
	}
	o.UserID = userID
	o.Status = domain.OrderPaid
	o.PaidAt = &now
	return nil
}

// deliver is the "paid -> delivered" step: the keys already listed on the
// order become sold, and the order becomes delivered (which also makes the
// API show their codes). keys are the order's keys as stored in the pool.
func deliver(o *domain.Order, keys []domain.Key, now time.Time) error {
	if o.Status == domain.OrderDelivered {
		return nil // already done (retried checkout)
	}
	if o.Status != domain.OrderPaid {
		return fmt.Errorf("order %s is %q, not paid", o.ID, o.Status)
	}
	byID := map[string]*domain.Key{}
	for i := range keys {
		byID[keys[i].ID] = &keys[i]
	}
	for _, it := range o.Items {
		for _, ik := range it.Keys {
			k := byID[ik.ID]
			if k == nil || k.Status != domain.KeyReserved || k.OrderID != o.ID {
				return errKeysMissing
			}
			k.Status = domain.KeySold
		}
	}
	o.Status = domain.OrderDelivered
	o.DeliveredAt = &now
	return nil
}
