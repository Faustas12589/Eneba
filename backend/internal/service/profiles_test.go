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

func TestMergeViews(t *testing.T) {
	cases := []struct {
		guest, account, want []string
	}{
		{nil, []string{"a", "b"}, []string{"a", "b"}},                     // nothing to merge
		{[]string{"x", "y"}, nil, []string{"x", "y"}},                     // new account takes the guest's order
		{[]string{"x", "a"}, []string{"a", "b"}, []string{"x", "a", "b"}}, // guest views first, no duplicate
		{[]string{"x", "y"}, []string{"a", "b"}, []string{"x", "y", "a"}}, // max 3: the account's oldest drops off
	}
	for _, c := range cases {
		if got := mergeViews(c.guest, c.account, 3); !reflect.DeepEqual(got, c.want) {
			t.Errorf("mergeViews(%v, %v) = %v, want %v", c.guest, c.account, got, c.want)
		}
	}
}

// Signing in moves the guest's views and orders to the account; the guest keeps nothing.
func TestMergeGuest(t *testing.T) {
	s, cart, orders := newProfilesSvc(t)
	ctx := context.Background()
	const guest, user = "guest:browser-1", "uid-1"

	_ = s.RecordView(ctx, user, "p002")  // viewed while signed in, earlier
	_ = s.RecordView(ctx, guest, "p003") // then as a guest
	if _, err := cart.SetQuantity(ctx, testCart, testOffer, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := orders.Checkout(ctx, guest, testCart); err != nil { // an older guest order
		t.Fatal(err)
	}

	moved, err := s.MergeGuest(ctx, user, guest)
	if err != nil || moved != 1 {
		t.Fatalf("merge: moved %d, err %v", moved, err)
	}
	got, _ := s.RecentlyViewed(ctx, user, 8)
	if len(got) != 2 || got[0].ID != "p003" || got[1].ID != "p002" {
		t.Fatalf("account views after merge: %v", got)
	}
	if left, _ := s.RecentlyViewed(ctx, guest, 8); len(left) != 0 {
		t.Fatalf("guest still has views: %v", left)
	}
	if pg, _ := orders.List(ctx, user, 1, 10); pg.Total != 1 {
		t.Fatalf("account has %d orders, want 1", pg.Total)
	}
	if pg, _ := orders.List(ctx, guest, 1, 10); pg.Total != 0 {
		t.Fatalf("guest still has %d orders", pg.Total)
	}
	if sum, _ := orders.Summary(ctx, user); sum.OrderCount != 1 || sum.TotalSpent <= 0 {
		t.Fatalf("summary: %+v", sum)
	}

	// Running it again (e.g. a second sign-in) changes nothing.
	if moved, err := s.MergeGuest(ctx, user, guest); err != nil || moved != 0 {
		t.Fatalf("second merge: moved %d, err %v", moved, err)
	}
}

func TestSetFavoriteRule(t *testing.T) {
	cases := []struct {
		name string
		list []string
		id   string
		on   bool
		want []string
	}{
		{"add goes in front", []string{"a"}, "b", true, []string{"b", "a"}},
		{"add twice keeps its place", []string{"a", "b"}, "b", true, []string{"a", "b"}},
		{"remove", []string{"a", "b", "c"}, "b", false, []string{"a", "c"}},
		{"remove missing", []string{"a"}, "x", false, []string{"a"}},
	}
	for _, c := range cases {
		got, err := setFavorite(c.list, c.id, c.on, 3)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, %v; want %v", c.name, got, err, c.want)
		}
	}
	if _, err := setFavorite([]string{"a", "b", "c"}, "d", true, 3); !errors.Is(err, ErrFavoritesFull) {
		t.Errorf("adding past the cap: err %v, want ErrFavoritesFull", err)
	}
	orig := []string{"a", "b"}
	_, _ = setFavorite(orig, "a", false, 3)
	if !reflect.DeepEqual(orig, []string{"a", "b"}) {
		t.Errorf("removing changed the caller's slice: %v", orig)
	}
}

// Favorites: newest first, safe to repeat, unknown products refused, per user,
// and a favorite counts as history for recommendations.
func TestFavorites(t *testing.T) {
	s, _, _ := newProfilesSvc(t)
	ctx := context.Background()
	const user = "uid-1"

	for _, id := range []string{"p002", "p003", "p003"} { // the second p003 changes nothing
		if err := s.SetFavorite(ctx, user, id, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetFavorite(ctx, user, "nope", true); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("unknown product: err %v", err)
	}
	got, _ := s.Favorites(ctx, user)
	if len(got) != 2 || got[0].ID != "p003" || got[1].ID != "p002" {
		t.Fatalf("favorites: %v", got)
	}
	if other, _ := s.Favorites(ctx, "uid-2"); len(other) != 0 {
		t.Fatalf("another user sees uid-1's favorites: %v", other)
	}
	rec, _ := s.Recommended(ctx, user, 50)
	for _, p := range rec {
		if p.ID == "p002" || p.ID == "p003" {
			t.Fatalf("recommended a favorite: %s", p.ID)
		}
	}

	if err := s.SetFavorite(ctx, user, "p003", false); err != nil {
		t.Fatal(err)
	}
	if err := s.SetFavorite(ctx, user, "p003", false); err != nil { // removing twice is fine
		t.Fatal(err)
	}
	if got, _ := s.Favorites(ctx, user); len(got) != 1 || got[0].ID != "p002" {
		t.Fatalf("after remove: %v", got)
	}
}

// Signing in moves a guest's favorites in front of the account's, without duplicates.
func TestMergeGuestFavorites(t *testing.T) {
	s, _, _ := newProfilesSvc(t)
	ctx := context.Background()
	const guest, user = "guest:browser-1", "uid-1"
	_ = s.SetFavorite(ctx, user, "p002", true)
	_ = s.SetFavorite(ctx, guest, "p002", true) // on both lists
	_ = s.SetFavorite(ctx, guest, "p004", true)

	if _, err := s.MergeGuest(ctx, user, guest); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Favorites(ctx, user)
	if len(got) != 2 || got[0].ID != "p004" || got[1].ID != "p002" {
		t.Fatalf("account favorites after merge: %v", got)
	}
	if left, _ := s.Favorites(ctx, guest); len(left) != 0 {
		t.Fatalf("guest still has favorites: %v", left)
	}
}
