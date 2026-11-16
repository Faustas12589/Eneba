package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

func newSvc(t *testing.T) *Catalog {
	t.Helper()
	m, err := repository.NewMemory()
	if err != nil {
		t.Fatal(err)
	}
	return NewCatalog(m)
}

func TestListFilterSortPaginate(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	pg, err := s.List(ctx, Filter{Platforms: []string{"Steam"}, Sort: "price_asc", PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	if pg.Total == 0 || len(pg.Items) > 5 {
		t.Fatalf("unexpected page: total=%d items=%d", pg.Total, len(pg.Items))
	}
	for i, p := range pg.Items {
		if p.Platform != "Steam" {
			t.Errorf("non-Steam product %q slipped through", p.Name)
		}
		if i > 0 && pg.Items[i-1].PriceFrom > p.PriceFrom {
			t.Errorf("not sorted ascending by price at %d", i)
		}
	}
	if len(pg.Facets.Platforms) < 2 {
		t.Errorf("facets should list all platforms, got %v", pg.Facets.Platforms)
	}

	capped, _ := s.List(ctx, Filter{MaxPrice: 1000})
	for _, p := range capped.Items {
		if p.PriceFrom > 1000 {
			t.Errorf("%q exceeds max price", p.Name)
		}
	}
}

func TestListByCategory(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	pg, err := s.List(ctx, Filter{Categories: []string{"dlc", "software"}})
	if err != nil {
		t.Fatal(err)
	}
	if pg.Total != 4 {
		t.Fatalf("want the 2 DLCs + 2 software products, got %d", pg.Total)
	}
	for _, p := range pg.Items {
		if p.Category != "dlc" && p.Category != "software" {
			t.Errorf("%q (%s) slipped through", p.Name, p.Category)
		}
	}
	// Combines with other filters: Steam DLC only.
	if steam, _ := s.List(ctx, Filter{Categories: []string{"dlc"}, Platforms: []string{"Steam"}}); steam.Total != 1 {
		t.Errorf("Steam DLC: got %d", steam.Total)
	}
	// Facets list every category in menu order, and no empty genre.
	if !reflect.DeepEqual(pg.Facets.Categories, domain.Categories) {
		t.Errorf("category facets %v", pg.Facets.Categories)
	}
	for _, g := range pg.Facets.Genres {
		if g == "" {
			t.Error("empty genre in facets")
		}
	}
}

// The price range covers the products matching every filter except price,
// so narrowing the price does not shrink the slider's own bounds.
func TestListPriceRange(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	pg, err := s.List(ctx, Filter{Categories: []string{"dlc"}})
	if err != nil {
		t.Fatal(err)
	}
	lo, hi := pg.Items[0].PriceFrom, pg.Items[0].PriceFrom
	for _, p := range pg.Items {
		lo, hi = min(lo, p.PriceFrom), max(hi, p.PriceFrom)
	}
	if pg.Facets.PriceMin != lo || pg.Facets.PriceMax != hi {
		t.Fatalf("DLC price range: got %d-%d, want %d-%d", pg.Facets.PriceMin, pg.Facets.PriceMax, lo, hi)
	}
	narrowed, _ := s.List(ctx, Filter{Categories: []string{"dlc"}, MinPrice: hi})
	if narrowed.Total != 1 && lo != hi {
		t.Errorf("min price = most expensive DLC: got %d products", narrowed.Total)
	}
	if narrowed.Facets.PriceMin != lo || narrowed.Facets.PriceMax != hi {
		t.Errorf("price filter changed the range: %d-%d", narrowed.Facets.PriceMin, narrowed.Facets.PriceMax)
	}
}

func TestListSearchTypo(t *testing.T) {
	s := newSvc(t)
	pg, err := s.List(context.Background(), Filter{Q: "eldne rign"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pg.Items) == 0 || pg.Items[0].Name != "Elden Ring" {
		t.Fatalf("typo search failed: %+v", pg.Items)
	}
}

func TestDiscountPercent(t *testing.T) {
	cases := []struct{ list, from, want int }{
		{5999, 3660, 39}, // 38.99% rounds to 39
		{2000, 1000, 50},
		{2999, 3139, 0}, // cheapest offer above list price: no discount
		{0, 100, 0},     // missing list price must not divide by zero
	}
	for _, c := range cases {
		if got := discountPercent(c.list, c.from); got != c.want {
			t.Errorf("discountPercent(%d, %d) = %d, want %d", c.list, c.from, got, c.want)
		}
	}
}

func TestGetSortsOffersCheapestFirst(t *testing.T) {
	s := newSvc(t)
	d, err := s.Get(context.Background(), "p001")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Offers) < 2 {
		t.Fatalf("expected several offers, got %d", len(d.Offers))
	}
	for i := 1; i < len(d.Offers); i++ {
		if d.Offers[i-1].Price > d.Offers[i].Price {
			t.Error("offers not sorted cheapest first")
		}
	}
	if _, err := s.Get(context.Background(), "nope"); err == nil {
		t.Error("expected not-found error")
	}
}

func TestIsPreOrder(t *testing.T) {
	now := time.Date(2026, 11, 15, 23, 59, 0, 0, time.Local)
	cases := map[string]bool{
		"":           false, // already out
		"2026-11-14": false,
		"2026-11-15": false, // release day: not a pre-order any more
		"2026-11-16": true,
		"2027-01-01": true,
	}
	for date, want := range cases {
		if got := isPreOrder(date, now); got != want {
			t.Errorf("isPreOrder(%q) = %v, want %v", date, got, want)
		}
	}
}

// The catalog-only homepage blocks: per-category best sellers, pre-orders, counts.
func TestHomeBlocks(t *testing.T) {
	s := newSvc(t)
	s.now = func() time.Time { return time.Date(2026, 11, 15, 12, 0, 0, 0, time.Local) } // seed has releases after this
	ctx := context.Background()

	topUps, _ := s.TopInCategory(ctx, "gaming-cards", 8)
	if len(topUps) == 0 {
		t.Fatal("no top-ups")
	}
	for i, p := range topUps {
		if p.Category != "gaming-cards" {
			t.Errorf("%s is not a gaming card", p.Name)
		}
		if i > 0 && topUps[i-1].Sales < p.Sales {
			t.Error("top-ups not sorted by sales")
		}
	}

	pre, _ := s.PreOrders(ctx, 8)
	if len(pre) == 0 {
		t.Fatal("seed has future release dates, expected pre-orders")
	}
	for i, p := range pre {
		if !p.PreOrder || p.ReleaseDate == "" {
			t.Errorf("%s listed as pre-order: %+v", p.Name, p)
		}
		if i > 0 && pre[i-1].ReleaseDate > p.ReleaseDate {
			t.Error("pre-orders not sorted soonest first")
		}
	}

	counts, _ := s.CategoryCounts(ctx)
	all, _ := s.List(ctx, Filter{})
	total := 0
	for i, c := range counts {
		if c.Category != domain.Categories[i] {
			t.Errorf("counts not in menu order: %v", counts)
		}
		total += c.Count
	}
	if total != all.Total {
		t.Errorf("category counts add up to %d, catalog has %d", total, all.Total)
	}
}
