package seeds

import (
	"slices"
	"testing"

	"eneba-marketplace/backend/internal/domain"
)

// Guards hand-edited seed data: every product has a known category and at
// least one offer, and priceFrom is really its cheapest offer.
func TestProductsMatchOffers(t *testing.T) {
	var products []domain.Product
	var offers []domain.Offer
	if err := Load("products", &products); err != nil {
		t.Fatal(err)
	}
	if err := Load("offers", &offers); err != nil {
		t.Fatal(err)
	}
	cheapest := map[string]int{}
	for _, o := range offers {
		if c, ok := cheapest[o.ProductID]; !ok || o.Price < c {
			cheapest[o.ProductID] = o.Price
		}
	}
	for _, p := range products {
		if !slices.Contains(domain.Categories, p.Category) {
			t.Errorf("%s: unknown category %q", p.ID, p.Category)
		}
		if c, ok := cheapest[p.ID]; !ok || c != p.PriceFrom {
			t.Errorf("%s: priceFrom %d, cheapest offer %d (has offers: %v)", p.ID, p.PriceFrom, c, ok)
		}
	}
}
