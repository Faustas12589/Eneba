package repository

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"eneba-marketplace/backend/internal/domain"
)

// Collection names. Keep in sync with cmd/seed.
const (
	ColProducts = "products"
	ColOffers   = "offers"
	ColBanners  = "banners"
)

type Firestore struct {
	c *firestore.Client
}

func NewFirestore(c *firestore.Client) *Firestore { return &Firestore{c: c} }

func (f *Firestore) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return listAll[domain.Product](ctx, f.c.Collection(ColProducts).Query)
}

func (f *Firestore) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	snap, err := f.c.Collection(ColProducts).Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var p domain.Product
	if err := snap.DataTo(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (f *Firestore) ListOffers(ctx context.Context, productID string) ([]domain.Offer, error) {
	return listAll[domain.Offer](ctx, f.c.Collection(ColOffers).Where("productId", "==", productID))
}

func (f *Firestore) ListBanners(ctx context.Context) ([]domain.Banner, error) {
	return listAll[domain.Banner](ctx, f.c.Collection(ColBanners).Query)
}

func listAll[T any](ctx context.Context, q firestore.Query) ([]T, error) {
	it := q.Documents(ctx)
	defer it.Stop()
	var out []T
	for {
		snap, err := it.Next()
		if errors.Is(err, iterator.Done) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		var v T
		if err := snap.DataTo(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
}
