package service

import (
	"context"
	"reflect"
	"testing"

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
