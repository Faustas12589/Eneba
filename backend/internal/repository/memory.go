package repository

import (
	"context"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/seeds"
)

// Memory serves the embedded seed data. Handy for local dev and tests
// before Firebase credentials are set up.
type Memory struct {
	products []domain.Product
	offers   []domain.Offer
	banners  []domain.Banner
}

func NewMemory() (*Memory, error) {
	m := &Memory{}
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
