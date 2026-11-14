package repository

import (
	"context"
	"sync"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/seeds"
)

// Memory serves the embedded seed data. Handy for local dev and tests
// before Firebase credentials are set up. Stock changes live until restart.
type Memory struct {
	products []domain.Product
	banners  []domain.Banner

	mu           sync.Mutex // guards offers (stock) and reservations
	offers       []domain.Offer
	reservations map[string]domain.Reservation // key: cartID + "_" + offerID
}

func NewMemory() (*Memory, error) {
	m := &Memory{reservations: map[string]domain.Reservation{}}
	if err := seeds.Load("products", &m.products); err != nil {
		return nil, err
	}
	if err := seeds.Load("offers", &m.offers); err != nil {
		return nil, err
	}
	if err := seeds.Load("banners", &m.banners); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Memory) ListProducts(context.Context) ([]domain.Product, error) {
	return append([]domain.Product(nil), m.products...), nil
}

func (m *Memory) GetProduct(_ context.Context, id string) (*domain.Product, error) {
	for i := range m.products {
		if m.products[i].ID == id {
			p := m.products[i]
			return &p, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListOffers(_ context.Context, productID string) ([]domain.Offer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Offer
	for _, o := range m.offers {
		if o.ProductID == productID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (m *Memory) ListBanners(context.Context) ([]domain.Banner, error) {
	return append([]domain.Banner(nil), m.banners...), nil
}

// UpdateReservation holds the mutex for the whole read-apply-write, which is
// the in-memory equivalent of a Firestore transaction.
func (m *Memory) UpdateReservation(_ context.Context, cartID, offerID string, apply ApplyFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := -1
	for i := range m.offers {
		if m.offers[i].ID == offerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrNotFound
	}

	key := reservationID(cartID, offerID)
	newStock, newReserved, err := apply(m.offers[idx].Stock, m.reservations[key].Quantity)
	if err != nil {
		return err
	}
	m.offers[idx].Stock = newStock
	if newReserved == 0 {
		delete(m.reservations, key)
	} else {
		m.reservations[key] = domain.Reservation{CartID: cartID, OfferID: offerID, Quantity: newReserved, UpdatedAt: time.Now()}
	}
	return nil
}

// reservationID is the document id: one reservation per (cart, offer) pair.
func reservationID(cartID, offerID string) string { return cartID + "_" + offerID }
