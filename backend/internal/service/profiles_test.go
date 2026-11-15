package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

func TestAddView(t *testing.T) {
	cases := []struct {
		list []string
		id   string
		want []string
	}{
		{nil, "a", []string{"a"}},
		{[]string{"a", "b"}, "c", []string{"c", "a", "b"}},
		{[]string{"a", "b", "c"}, "b", []string{"b", "a", "c"}}, // seen again: moves to the front, no duplicate
		{[]string{"a", "b", "c"}, "d", []string{"d", "a", "b"}}, // max 3: the oldest drops off
	}
	for _, c := range cases {
		if got := addView(c.list, c.id, 3); !reflect.DeepEqual(got, c.want) {
			t.Errorf("addView(%v, %q) = %v, want %v", c.list, c.id, got, c.want)
		}
	}
}

func TestRecommendScoresByHistory(t *testing.T) {
	p := func(id, genre, category, platform string, sales int) domain.Product {
		return domain.Product{ID: id, Genre: genre, Category: category, Platform: platform, Sales: sales}
	}
	viewedRPG := p("rpg1", "RPG", "games", "Steam", 10)
	all := []domain.Product{
		viewedRPG,
		p("card", "", "gift-cards", "Amazon", 999),         // best seller, but nothing in common
		p("rpg2", "RPG", "games", "GOG", 5),                // genre + category = 5
		p("steamShooter", "Shooter", "games", "Steam", 50), // category + platform = 3
		p("steamCard", "", "gaming-cards", "Steam", 60),    // platform = 1
	}
	got := recommend(all, []domain.Product{viewedRPG}, 10)
	var ids []string
	for _, r := range got {
		ids = append(ids, r.ID)
	}
	want := []string{"rpg2", "steamShooter", "steamCard", "card"} // viewed product left out
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("got %v, want %v", ids, want)
	}

	// No history: plain best sellers.
	if top := recommend(all, nil, 1); top[0].ID != "card" {
		t.Fatalf("no history should give best sellers first, got %s", top[0].ID)
	}
	// Cards and software have no genre: "" must not count as a shared genre.
	// If it did, the low-selling key would score 3 and jump above the game.
	bought := p("giftCard", "", "gift-cards", "Amazon", 0)
	noGenre := []domain.Product{bought, p("game", "RPG", "games", "GOG", 50), p("softwareKey", "", "software", "Microsoft", 1)}
	if got := recommend(noGenre, []domain.Product{bought}, 1); got[0].ID != "game" {
		t.Fatalf("empty genre matched: top is %s", got[0].ID)
	}
}

func newProfilesSvc(t *testing.T) (*Profiles, *Cart, *Orders) {
	t.Helper()
	m, err := repository.NewMemory()
	if err != nil {
		t.Fatal(err)
	}
	return NewProfiles(m, m, NewCatalog(m)), NewCart(m), NewOrders(m)
}

func TestRecentlyViewedPerUser(t *testing.T) {
	s, _, _ := newProfilesSvc(t)
	ctx := context.Background()

	if got, err := s.RecentlyViewed(ctx, "user-1", 8); err != nil || len(got) != 0 {
		t.Fatalf("new user: %v %v", got, err)
	}
	for _, id := range []string{"p001", "p002", "p001", "p003"} {
		if err := s.RecordView(ctx, "user-1", id); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := s.RecentlyViewed(ctx, "user-1", 8)
	var ids []string
	for _, p := range got {
		ids = append(ids, p.ID)
	}
	if !reflect.DeepEqual(ids, []string{"p003", "p001", "p002"}) {
		t.Fatalf("recently viewed: %v", ids)
	}
	if other, _ := s.RecentlyViewed(ctx, "user-2", 8); len(other) != 0 {
		t.Fatal("another user sees user-1's views")
	}
	if err := s.RecordView(ctx, "user-1", "nope"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("unknown product: got %v", err)
	}
}

// Bought products count as history: they are left out of recommendations
// and similar products come first.
func TestRecommendedUsesPurchases(t *testing.T) {
	s, cart, orders := newProfilesSvc(t)
	ctx := context.Background()
	if _, err := cart.SetQuantity(ctx, testCart, testOffer, 1); err != nil { // p001 Elden Ring (RPG, Steam)
		t.Fatal(err)
	}
	if _, err := orders.Checkout(ctx, "user-1", testCart); err != nil {
		t.Fatal(err)
	}
	got, err := s.Recommended(ctx, "user-1", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 8 {
		t.Fatalf("want 8 recommendations, got %d", len(got))
	}
	for _, p := range got {
		if p.ID == "p001" {
			t.Fatal("recommended a product the user already bought")
		}
	}
	if got[0].Genre != "RPG" {
		t.Errorf("an RPG buyer should get an RPG first, got %s (%s)", got[0].Name, got[0].Genre)
	}
}
