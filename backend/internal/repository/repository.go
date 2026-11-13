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
