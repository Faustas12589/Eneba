package seeds

import (
	"testing"

	"eneba-marketplace/backend/internal/domain"
)

func TestKeysMatchStock(t *testing.T) {
	var offers []domain.Offer
	if err := Load("offers", &offers); err != nil {
		t.Fatal(err)
	}
	keys := Keys(offers)
	perOffer := map[string]int{}
	codes := map[string]bool{}
	for _, k := range keys {
		perOffer[k.OfferID]++
		codes[k.Code] = true
		if k.Status != domain.KeyAvailable || k.OrderID != "" {
			t.Fatalf("seed key not available: %+v", k)
		}
	}
	for _, o := range offers {
		if perOffer[o.ID] != o.Stock {
			t.Errorf("%s: %d keys, stock %d", o.ID, perOffer[o.ID], o.Stock)
		}
	}
	if len(codes) != len(keys) {
		t.Errorf("duplicate codes: %d unique of %d", len(codes), len(keys))
	}
}
