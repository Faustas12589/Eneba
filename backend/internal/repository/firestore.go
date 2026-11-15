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
	ColKeys     = "keys"
	ColOrders   = "orders" // carts are orders with status "cart"
	ColProfiles = "profiles" // doc id = user id (guest:<id> while login is off)

	// ColReservations held cart lines before they moved into orders.
	// Only cmd/seed still uses it, to delete leftovers.
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
	return decode[domain.Product](f.c.Collection(ColProducts).Doc(id).Get(ctx))
}

func (f *Firestore) ListOffers(ctx context.Context, productID string) ([]domain.Offer, error) {
	return listAll[domain.Offer](ctx, f.c.Collection(ColOffers).Where("productId", "==", productID))
}

func (f *Firestore) ListBanners(ctx context.Context) ([]domain.Banner, error) {
	return listAll[domain.Banner](ctx, f.c.Collection(ColBanners).Query)
}

// UpdateReservation runs in a Firestore transaction: if another request changes
// the offer, the cart or one of the picked keys between our read and write,
// Firestore retries the whole function with fresh data, so two carts can never
// take the same last unit or the same key.
func (f *Firestore) UpdateReservation(ctx context.Context, cartID, offerID string, apply ApplyFunc) (*domain.Order, error) {
	offerRef := f.c.Collection(ColOffers).Doc(offerID)
	cartRef := f.c.Collection(ColOrders).Doc(cartID)

	var out *domain.Order
	err := f.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Firestore transactions must do all reads before any write.
		offer, err := decode[domain.Offer](tx.Get(offerRef))
		if err != nil {
			return err
		}
		cart, err := decode[domain.Order](tx.Get(cartRef))
		switch {
		case errors.Is(err, ErrNotFound):
			cart = &domain.Order{ID: cartID, Status: domain.OrderCart, Currency: offer.Currency, CreatedAt: time.Now()}
		case err != nil:
			return err
		case cart.Status != domain.OrderCart:
			return ErrCartClosed
		}

		cur := cart.Item(offerID)
		before := reservationOf(cartID, cur)
		newStock, next, err := apply(offer.Stock, before)
		if err != nil {
			return err
		}
		need := next.Quantity - before.Quantity // keys to reserve; < 0 means some go back

		// A new line is written whole, so it needs the product's name and image.
		var product domain.Product
		if cur == nil && next.Quantity > 0 {
			p, err := decode[domain.Product](tx.Get(f.c.Collection(ColProducts).Doc(offer.ProductID)))
			if err != nil {
				return err
			}
			product = *p
		}

		var added []domain.Key
		var addedRefs []*firestore.DocumentRef
		if need > 0 {
			// Equality filters only, so Firestore needs no composite index.
			q := f.c.Collection(ColKeys).
				Where("offerId", "==", offerID).
				Where("status", "==", domain.KeyAvailable).
				Limit(need)
			snaps, err := tx.Documents(q).GetAll()
			if err != nil {
				return err
			}
			if len(snaps) < need {
				return errKeyPoolMismatch
			}
			for _, s := range snaps {
				var k domain.Key
				if err := s.DataTo(&k); err != nil {
					return err
				}
				added = append(added, k)
				addedRefs = append(addedRefs, s.Ref)
			}
		}

		item, released := nextItem(cur, *offer, product, next, added)
		for _, ref := range addedRefs {
			if err := setKey(tx, ref, domain.KeyReserved, cartID); err != nil {
				return err
			}
		}
		for _, k := range released {
			if err := setKey(tx, f.c.Collection(ColKeys).Doc(k.ID), domain.KeyAvailable, ""); err != nil {
				return err
			}
		}
		if err := tx.Update(offerRef, []firestore.Update{{Path: "stock", Value: newStock}}); err != nil {
			return err
		}
		cart.PutItem(item)
		out = cart
		if len(cart.Items) == 0 {
			return tx.Delete(cartRef)
		}
		return tx.Set(cartRef, cart)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func setKey(tx *firestore.Transaction, ref *firestore.DocumentRef, status, orderID string) error {
	return tx.Update(ref, []firestore.Update{{Path: "status", Value: status}, {Path: "orderId", Value: orderID}})
}

// Lists every open cart and keeps the expired lines in Go: the number of open
// carts is small, and a status + time query would need a composite index.
func (f *Firestore) ListExpiredReservations(ctx context.Context, now time.Time) ([]domain.Reservation, error) {
	carts, err := listAll[domain.Order](ctx, f.c.Collection(ColOrders).Where("status", "==", domain.OrderCart))
	if err != nil {
		return nil, err
	}
	var out []domain.Reservation
	for _, o := range carts {
		out = append(out, expiredLines(o, now)...)
	}
	return out, nil
}

func (f *Firestore) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return decode[domain.Order](f.c.Collection(ColOrders).Doc(id).Get(ctx))
}

func (f *Firestore) PayOrder(ctx context.Context, id string, apply PayFunc) (*domain.Order, error) {
	ref := f.c.Collection(ColOrders).Doc(id)
	var out *domain.Order
	err := f.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		o, err := decode[domain.Order](tx.Get(ref))
		if err != nil {
			return err
		}
		if err := apply(o); err != nil {
			return err
		}
		out = o
		return tx.Set(ref, o)
	})
	return out, err
}

func (f *Firestore) DeliverOrder(ctx context.Context, id string, apply DeliverFunc) (*domain.Order, error) {
	ref := f.c.Collection(ColOrders).Doc(id)
	var out *domain.Order
	err := f.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		o, err := decode[domain.Order](tx.Get(ref))
		if err != nil {
			return err
		}
		var refs []*firestore.DocumentRef // the keys listed on the order's lines
		for _, it := range o.Items {
			for _, k := range it.Keys {
				refs = append(refs, f.c.Collection(ColKeys).Doc(k.ID))
			}
		}
		snaps, err := tx.GetAll(refs)
		if err != nil {
			return err
		}
		var keys []domain.Key
		for _, s := range snaps {
			if !s.Exists() {
				continue // the service notices the missing key
			}
			var k domain.Key
			if err := s.DataTo(&k); err != nil {
				return err
			}
			keys = append(keys, k)
		}
		if err := apply(o, keys); err != nil {
			return err
		}
		for _, k := range keys {
			if err := setKey(tx, f.c.Collection(ColKeys).Doc(k.ID), k.Status, k.OrderID); err != nil {
				return err
			}
		}
		out = o
		return tx.Set(ref, o)
	})
	return out, err
}

func (f *Firestore) ListUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	return listAll[domain.Order](ctx, f.c.Collection(ColOrders).Where("userId", "==", userID))
}

// decode turns a document read into a value; a missing document is ErrNotFound.
func decode[T any](snap *firestore.DocumentSnapshot, err error) (*T, error) {
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var v T
	if err := snap.DataTo(&v); err != nil {
		return nil, err
	}
	return &v, nil
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

func (f *Firestore) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return decode[domain.Profile](f.c.Collection(ColProfiles).Doc(userID).Get(ctx))
}

func (f *Firestore) UpdateProfile(ctx context.Context, userID string, apply ProfileFunc) (*domain.Profile, error) {
	ref := f.c.Collection(ColProfiles).Doc(userID)
	var out *domain.Profile
	err := f.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		p, err := decode[domain.Profile](tx.Get(ref))
		if errors.Is(err, ErrNotFound) {
			p, err = &domain.Profile{UserID: userID}, nil
		}
		if err != nil {
			return err
		}
		if err := apply(p); err != nil {
			return err
		}
		out = p
		return tx.Set(ref, p)
	})
	return out, err
}
