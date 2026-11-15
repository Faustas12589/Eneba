package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

const otherOffer = "p001-o2" // GameVault, 36.60 EUR

// checkKeys fails unless the offer's stock equals its available keys and
// `reserved` / `sold` keys are in those states.
func checkKeys(t *testing.T, m *repository.Memory, offerID string, reserved, sold int) {
	t.Helper()
	c := m.KeyCounts(offerID)
	if c[domain.KeyAvailable] != stockOf(t, m, offerID) || c[domain.KeyReserved] != reserved || c[domain.KeySold] != sold {
		t.Fatalf("%s: keys %v, stock %d; want reserved=%d sold=%d", offerID, c, stockOf(t, m, offerID), reserved, sold)
	}
}

func TestCartReservesKeys(t *testing.T) {
	s, m := newCartSvc(t)
	ctx := context.Background()
	checkKeys(t, m, testOffer, 0, 0) // seed: one available key per unit of stock

	for _, step := range []struct{ qty, reserved int }{{3, 3}, {5, 5}, {1, 1}, {0, 0}} {
		if _, err := s.SetQuantity(ctx, testCart, testOffer, step.qty); err != nil {
			t.Fatal(err)
		}
		checkKeys(t, m, testOffer, step.reserved, 0)
	}
	if _, err := m.GetOrder(ctx, testCart); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("an empty cart should not be stored, got %v", err)
	}
}

// A line is written whole when it goes into the cart; later changes only
// touch quantity, keys, expiry and total.
func TestCartLineWrittenOnce(t *testing.T) {
	s, m := newCartSvc(t)
	ctx := context.Background()
	res, err := s.SetQuantity(ctx, testCart, testOffer, 2)
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := m.GetOrder(ctx, testCart)
	it := stored.Items[0]
	if it.ProductName != "Elden Ring" || it.SellerName != "DigitalDragon" || it.Price != 4021 ||
		it.ExpiresAt == nil || len(it.Keys) != 2 || it.Keys[0].Code == "" || stored.Total != 2*4021 {
		t.Fatalf("cart line not written whole: %+v total=%d", it, stored.Total)
	}
	if res.Cart.Items[0].Keys != nil {
		t.Fatal("the API must not show key codes of an unpaid cart")
	}
	firstKey := it.Keys[0]

	if _, err := s.SetQuantity(ctx, testCart, testOffer, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetQuantity(ctx, testCart, testOffer, 1); err != nil {
		t.Fatal(err)
	}
	stored, _ = m.GetOrder(ctx, testCart)
	it = stored.Items[0]
	if len(it.Keys) != 1 || it.Keys[0] != firstKey || stored.Total != 4021 || it.ProductName != "Elden Ring" {
		t.Fatalf("existing line should keep its data and first key: %+v total=%d", it, stored.Total)
	}
}

func TestCheckoutOnlyChangesState(t *testing.T) {
	cart, m := newCartSvc(t)
	orders := NewOrders(m)
	ctx := context.Background()
	if _, err := cart.SetQuantity(ctx, testCart, testOffer, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := cart.SetQuantity(ctx, testCart, otherOffer, 1); err != nil {
		t.Fatal(err)
	}
	before, _ := m.GetOrder(ctx, testCart)
	stockBefore := stockOf(t, m, testOffer)

	o, err := orders.Checkout(ctx, "user-1", testCart)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != domain.OrderDelivered || o.PaidAt == nil || o.DeliveredAt == nil {
		t.Fatalf("want delivered order with both dates, got %+v", o)
	}

	// Same lines, same keys, same total: only the status fields changed.
	after, _ := m.GetOrder(ctx, testCart)
	want := before.Items
	for i := range want {
		want[i].ExpiresAt = nil // a paid line no longer expires
	}
	if !reflect.DeepEqual(after.Items, want) || after.Total != before.Total || after.Total != 2*4021+3660 {
		t.Fatalf("checkout changed line data:\nbefore %+v\nafter  %+v", want, after.Items)
	}
	if len(o.Items[0].Keys) != 2 || o.Items[0].Keys[0].Code == "" {
		t.Fatalf("a delivered order shows its codes: %+v", o.Items[0])
	}

	// Keys are sold; stock was already taken when they went into the cart.
	checkKeys(t, m, testOffer, 0, 2)
	checkKeys(t, m, otherOffer, 0, 1)
	if stockOf(t, m, testOffer) != stockBefore {
		t.Fatal("checkout must not change stock again")
	}

	// The paid cart is closed.
	if _, err := cart.SetQuantity(ctx, testCart, testOffer, 3); !errors.Is(err, repository.ErrCartClosed) {
		t.Fatalf("changing a paid cart: got %v", err)
	}
	if _, err := cart.Get(ctx, testCart); !errors.Is(err, repository.ErrCartClosed) {
		t.Fatalf("reading a paid cart as a cart: got %v", err)
	}
	if list, _ := orders.List(ctx, "user-1"); len(list) != 1 || list[0].ID != testCart {
		t.Fatalf("order history: %+v", list)
	}
	if list, _ := orders.List(ctx, "user-2"); len(list) != 0 {
		t.Fatalf("other user sees orders: %+v", list)
	}
}

func TestCheckoutRetryAndOwner(t *testing.T) {
	cart, m := newCartSvc(t)
	orders := NewOrders(m)
	ctx := context.Background()
	if _, err := cart.SetQuantity(ctx, testCart, testOffer, 1); err != nil {
		t.Fatal(err)
	}
	first, err := orders.Checkout(ctx, "user-1", testCart)
	if err != nil {
		t.Fatal(err)
	}
	again, err := orders.Checkout(ctx, "user-1", testCart)
	if err != nil || !reflect.DeepEqual(again, first) {
		t.Fatalf("retry should return the same order: %v\n%+v\n%+v", err, again, first)
	}
	checkKeys(t, m, testOffer, 0, 1)
	if _, err := orders.Checkout(ctx, "user-2", testCart); !errors.Is(err, ErrNotYourOrder) {
		t.Fatalf("someone else's cart: got %v", err)
	}
}

func TestCheckoutEmptyOrInvalid(t *testing.T) {
	_, m := newCartSvc(t)
	orders := NewOrders(m)
	ctx := context.Background()
	if _, err := orders.Checkout(ctx, "user-1", testCart); !errors.Is(err, ErrEmptyCart) {
		t.Errorf("empty cart: got %v", err)
	}
	if _, err := orders.Checkout(ctx, "user-1", "short"); !errors.Is(err, ErrInvalidCartID) {
		t.Errorf("bad cart id: got %v", err)
	}
}

func TestDeliverNeedsReservedKeys(t *testing.T) {
	o := &domain.Order{ID: "o1", Status: domain.OrderPaid, Items: []domain.OrderItem{
		{OfferID: "a", Quantity: 2, Keys: []domain.ItemKey{{ID: "k1"}, {ID: "k2"}}},
	}}
	keys := []domain.Key{{ID: "k1", Status: domain.KeyReserved, OrderID: "o1"}} // k2 missing
	if err := deliver(o, keys, time.Now()); !errors.Is(err, errKeysMissing) {
		t.Fatalf("got %v", err)
	}
	if o.Status != domain.OrderPaid || o.DeliveredAt != nil {
		t.Fatal("a failed delivery must leave the order paid")
	}
}

func TestPayEmptyCart(t *testing.T) {
	o := &domain.Order{Status: domain.OrderCart}
	if err := pay(o, "u", time.Now()); !errors.Is(err, ErrEmptyCart) || o.PaidAt != nil {
		t.Fatalf("empty: got %v, paidAt %v", err, o.PaidAt)
	}
}
