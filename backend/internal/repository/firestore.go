package repository

import (
	"context"
	"errors"
	"time"

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

	ColReservations = "reservations"
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

// UpdateReservation runs in a Firestore transaction: if another request changes
// the offer's stock between our read and write, Firestore retries the whole
// function with fresh data, so two carts can never take the same last unit.
func (f *Firestore) UpdateReservation(ctx context.Context, cartID, offerID string, apply ApplyFunc) error {
	offerRef := f.c.Collection(ColOffers).Doc(offerID)
	resRef := f.c.Collection(ColReservations).Doc(reservationID(cartID, offerID))

	return f.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		offerSnap, err := tx.Get(offerRef)
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		var offer domain.Offer
		if err := offerSnap.DataTo(&offer); err != nil {
			return err
		}

		var res domain.Reservation // zero Quantity if this cart has none yet
		resSnap, err := tx.Get(resRef)
		if err != nil && status.Code(err) != codes.NotFound {
			return err
		}
		if err == nil {
			if err := resSnap.DataTo(&res); err != nil {
				return err
			}
		}

		newStock, newReserved, err := apply(offer.Stock, res.Quantity)
		if err != nil {
			return err
		}
		if err := tx.Update(offerRef, []firestore.Update{{Path: "stock", Value: newStock}}); err != nil {
			return err
		}
		if newReserved == 0 {
			return tx.Delete(resRef)
		}
		return tx.Set(resRef, domain.Reservation{
			CartID: cartID, OfferID: offerID, Quantity: newReserved, UpdatedAt: time.Now(),
		})
	})
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
