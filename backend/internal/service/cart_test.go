package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"eneba-marketplace/backend/internal/repository"
)

const (
	testCart  = "cart-aaaaaaaaaaaaaaaa"
	testOffer = "p001-o1" // seeded with stock 7
)

func newCartSvc(t *testing.T) (*Cart, *repository.Memory) {
	t.Helper()
	m, err := repository.NewMemory()
	if err != nil {
		t.Fatal(err)
	}
	return NewCart(m), m
}

func stockOf(t *testing.T, m *repository.Memory, offerID string) int {
	t.Helper()
	offers, _ := m.ListOffers(context.Background(), "p001")
	for _, o := range offers {
		if o.ID == offerID {
			return o.Stock
		}
	}
	t.Fatalf("offer %s not found", offerID)
	return 0
}

func TestReserve(t *testing.T) {
	cases := []struct {
		stock, reserved, want, newStock int
		wantErr                         bool
	}{
		{stock: 5, reserved: 0, want: 2, newStock: 3},
		{stock: 3, reserved: 2, want: 5, newStock: 0},   // takes the last 3
		{stock: 3, reserved: 2, want: 6, wantErr: true}, // needs 4, only 3
		{stock: 0, reserved: 4, want: 1, newStock: 3},   // gives 3 back
		{stock: 0, reserved: 0, want: 1, wantErr: true}, // sold out
	}
	for _, c := range cases {
		got, err := reserve(c.stock, c.reserved, c.want)
		if (err != nil) != c.wantErr || (!c.wantErr && got != c.newStock) {
			t.Errorf("reserve(%d,%d,%d) = %d, %v", c.stock, c.reserved, c.want, got, err)
		}
	}
}

func TestSetQuantityMovesStock(t *testing.T) {
	s, m := newCartSvc(t)
	ctx := context.Background()
	start := stockOf(t, m, testOffer)

	res, err := s.SetQuantity(ctx, testCart, testOffer, 3)
	if err != nil {
		t.Fatal(err)
	}
	if it := res.Cart.Item(testOffer); it == nil || it.Quantity != 3 || res.Stock != start-3 || stockOf(t, m, testOffer) != start-3 {
		t.Fatalf("after reserving 3: res=%+v stock=%d", res, stockOf(t, m, testOffer))
	}

	// Same request again (e.g. a retry) changes nothing.
	if _, err := s.SetQuantity(ctx, testCart, testOffer, 3); err != nil || stockOf(t, m, testOffer) != start-3 {
		t.Fatalf("retry changed stock: %v %d", err, stockOf(t, m, testOffer))
	}

	// Removing the line gives everything back.
	if _, err := s.SetQuantity(ctx, testCart, testOffer, 0); err != nil || stockOf(t, m, testOffer) != start {
		t.Fatalf("remove did not restore stock: %v %d", err, stockOf(t, m, testOffer))
	}
}

func TestSetQuantityOutOfStock(t *testing.T) {
	s, m := newCartSvc(t)
	ctx := context.Background()

	// Another cart drains the offer down to 2 units, MaxPerLine at a time.
	for i := 0; stockOf(t, m, testOffer) > 2; i++ {
		take := min(MaxPerLine, stockOf(t, m, testOffer)-2)
		cart := fmt.Sprintf("drain-cart-%06d", i)
		if _, err := s.SetQuantity(ctx, cart, testOffer, take); err != nil {
			t.Fatal(err)
		}
	}

	_, err := s.SetQuantity(ctx, testCart, testOffer, 3)
	var oos *OutOfStockError
	if !errors.As(err, &oos) || oos.Available != 2 {
		t.Fatalf("want OutOfStockError{Available: 2}, got %v", err)
	}
	if stockOf(t, m, testOffer) != 2 {
		t.Fatal("a failed reservation must not change stock")
	}
}

func TestSetQuantityValidation(t *testing.T) {
	s, _ := newCartSvc(t)
	ctx := context.Background()
	if _, err := s.SetQuantity(ctx, "short", testOffer, 1); !errors.Is(err, ErrInvalidCartID) {
		t.Errorf("short cart id: got %v", err)
	}
	if _, err := s.SetQuantity(ctx, testCart, testOffer, MaxPerLine+1); !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("too many: got %v", err)
	}
	if _, err := s.SetQuantity(ctx, testCart, "no-such-offer", 1); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("unknown offer: got %v", err)
	}
}

// Many carts racing for the last units: exactly `stock` of them may win.
func TestSetQuantityNoOversell(t *testing.T) {
	s, m := newCartSvc(t)
	ctx := context.Background()
	stock := stockOf(t, m, testOffer)
	buyers := stock + 50

	var wg sync.WaitGroup
	var mu sync.Mutex
	won := 0
	for i := range buyers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cart := fmt.Sprintf("race-cart-%06d", i)
			if _, err := s.SetQuantity(ctx, cart, testOffer, 1); err == nil {
				mu.Lock()
				won++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if won != stock || stockOf(t, m, testOffer) != 0 {
		t.Fatalf("won=%d (want %d), stock left=%d", won, stock, stockOf(t, m, testOffer))
	}
	checkKeys(t, m, testOffer, stock, 0) // every winner holds exactly one key
}

func TestReleaseExpired(t *testing.T) {
	s, m := newCartSvc(t)
	ctx := context.Background()
	start := stockOf(t, m, testOffer)
	if _, err := s.SetQuantity(ctx, testCart, testOffer, 4); err != nil {
		t.Fatal(err)
	}

	// Not expired yet: nothing happens.
	if n, err := s.ReleaseExpired(ctx, time.Now()); err != nil || n != 0 {
		t.Fatalf("released %d too early (err %v)", n, err)
	}
	if c, _ := s.Get(ctx, testCart); len(c.Items) != 1 || c.Items[0].Quantity != 4 {
		t.Fatalf("cart should still hold 4: %+v", c)
	}

	// Pretend 31 minutes passed: stock comes back and the cart is empty.
	n, err := s.ReleaseExpired(ctx, time.Now().Add(ReservationTTL+time.Minute))
	if err != nil || n != 1 {
		t.Fatalf("want 1 released, got %d (err %v)", n, err)
	}
	if stockOf(t, m, testOffer) != start {
		t.Fatalf("stock not restored: %d, want %d", stockOf(t, m, testOffer), start)
	}
	if c, _ := s.Get(ctx, testCart); len(c.Items) != 0 {
		t.Fatalf("expired line still in cart: %+v", c)
	}
	checkKeys(t, m, testOffer, 0, 0) // reserved keys are available again
}
