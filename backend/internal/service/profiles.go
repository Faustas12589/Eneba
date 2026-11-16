package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

// maxViewed is how many recently viewed products a profile keeps.
const maxViewed = 20

// MaxFavorites caps the wishlist. Adding past it is refused (ErrFavoritesFull)
// rather than silently dropping an old favorite, which a shopper would not expect.
const MaxFavorites = 100

var ErrFavoritesFull = fmt.Errorf("favorites list is full (max %d)", MaxFavorites)

// Profiles powers the personal homepage blocks: "Neseniai peržiūrėta" and
// "Rekomenduojama tau" (DECISIONS #25). The user id comes from the request
// (a guest id while login is off), so this works the same once login is back.
type Profiles struct {
	repo    repository.Profiles
	orders  repository.Orders
	catalog *Catalog
}

func NewProfiles(repo repository.Profiles, orders repository.Orders, catalog *Catalog) *Profiles {
	return &Profiles{repo: repo, orders: orders, catalog: catalog}
}

// RecordView puts a product at the front of the user's recently viewed list.
// Unknown products return repository.ErrNotFound, so the list only holds real ids.
func (s *Profiles) RecordView(ctx context.Context, userID, productID string) error {
	found, err := s.catalog.lookup(ctx, []string{productID})
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return repository.ErrNotFound
	}
	// Reloading the same product page changes nothing, so skip the write.
	if p, err := s.repo.GetProfile(ctx, userID); err == nil && len(p.RecentlyViewed) > 0 && p.RecentlyViewed[0] == productID {
		return nil
	}
	now := time.Now()
	_, err = s.repo.UpdateProfile(ctx, userID, func(p *domain.Profile) error {
		p.RecentlyViewed = addView(p.RecentlyViewed, productID, maxViewed)
		p.UpdatedAt = now
		return nil
	})
	return err
}

// RecentlyViewed returns up to n of the user's viewed products, newest first.
func (s *Profiles) RecentlyViewed(ctx context.Context, userID string, n int) ([]domain.Product, error) {
	viewed, err := s.viewed(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(viewed) > n {
		viewed = viewed[:n]
	}
	return s.catalog.lookup(ctx, viewed)
}

// Recommended returns n products the user has neither viewed nor bought,
// ranked by how much they share with what they did view or buy.
func (s *Profiles) Recommended(ctx context.Context, userID string, n int) ([]domain.Product, error) {
	viewed, err := s.viewed(ctx, userID)
	if err != nil {
		return nil, err
	}
	orders, err := s.orders.ListUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	favs, err := s.favorites(ctx, userID)
	if err != nil {
		return nil, err
	}
	historyIDs := append(slices.Clone(viewed), favs...) // a favorite says even more than a view
	for _, o := range orders {
		for _, it := range o.Items {
			historyIDs = append(historyIDs, it.ProductID)
		}
	}
	history, err := s.catalog.lookup(ctx, historyIDs)
	if err != nil {
		return nil, err
	}
	all, err := s.catalog.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return recommend(all, history, n), nil
}

// MergeGuest runs when a guest signs in: the guest's recently viewed products
// and favorites go to the front of the account's lists, and the guest's orders
// become the account's (DECISIONS #28, #37). Safe to repeat: lists are deduplicated
// and an order that already moved is no longer the guest's. Returns how many orders moved.
func (s *Profiles) MergeGuest(ctx context.Context, userID, guestUserID string) (int, error) {
	guest, err := s.repo.GetProfile(ctx, guestUserID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return 0, err
	}
	if guest != nil && (len(guest.RecentlyViewed) > 0 || len(guest.Favorites) > 0) {
		now := time.Now()
		if _, err := s.repo.UpdateProfile(ctx, userID, func(p *domain.Profile) error {
			p.RecentlyViewed = mergeViews(guest.RecentlyViewed, p.RecentlyViewed, maxViewed)
			// a merge never drops a favorite, so the cap may be passed here once
			p.Favorites = mergeViews(guest.Favorites, p.Favorites, len(guest.Favorites)+len(p.Favorites))
			p.UpdatedAt = now
			return nil
		}); err != nil {
			return 0, err
		}
		// Empty the guest's lists only after the account has them: a crash in between
		// leaves a duplicate, which the next merge removes, never a lost item.
		if _, err := s.repo.UpdateProfile(ctx, guestUserID, func(p *domain.Profile) error {
			p.RecentlyViewed, p.Favorites, p.UpdatedAt = nil, nil, now
			return nil
		}); err != nil {
			return 0, err
		}
	}
	return s.orders.MoveUserOrders(ctx, guestUserID, userID)
}

// SetFavorite adds a product to the user's favorites (on) or removes it. Absolute,
// like the cart's quantity (#15): sending it twice changes nothing, so a retried
// request or a double click is harmless. Unknown products return ErrNotFound.
func (s *Profiles) SetFavorite(ctx context.Context, userID, productID string, on bool) error {
	if on {
		found, err := s.catalog.lookup(ctx, []string{productID})
		if err != nil {
			return err
		}
		if len(found) == 0 {
			return repository.ErrNotFound
		}
	}
	// Already as asked: skip the write (Firestore quota, like RecordView).
	favs, err := s.favorites(ctx, userID)
	if err != nil {
		return err
	}
	if slices.Contains(favs, productID) == on {
		return nil
	}
	now := time.Now()
	_, err = s.repo.UpdateProfile(ctx, userID, func(p *domain.Profile) error {
		next, err := setFavorite(p.Favorites, productID, on, MaxFavorites)
		if err != nil {
			return err
		}
		p.Favorites, p.UpdatedAt = next, now
		return nil
	})
	return err
}

// Favorites returns the user's favorite products, newest first. Products that
// left the catalog are skipped (lookup ignores unknown ids).
func (s *Profiles) Favorites(ctx context.Context, userID string) ([]domain.Product, error) {
	favs, err := s.favorites(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.catalog.lookup(ctx, favs)
}

// favorites is the user's favorite ids; a user without a profile has none.
func (s *Profiles) favorites(ctx context.Context, userID string) ([]string, error) {
	p, err := s.repo.GetProfile(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p.Favorites, nil
}

// setFavorite returns the list with id added at the front (on) or removed.
// Adding one that is already there keeps its place; adding past max is an error.
func setFavorite(list []string, id string, on bool, max int) ([]string, error) {
	i := slices.Index(list, id)
	switch {
	case on && i >= 0, !on && i < 0: // already as asked
		return list, nil
	case on:
		if len(list) >= max {
			return nil, ErrFavoritesFull
		}
		return append([]string{id}, list...), nil
	default:
		return slices.Delete(slices.Clone(list), i, i+1), nil
	}
}

// viewed is the user's recently viewed ids; a user without a profile has none.
func (s *Profiles) viewed(ctx context.Context, userID string) ([]string, error) {
	p, err := s.repo.GetProfile(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p.RecentlyViewed, nil
}

// addView moves id to the front of the list (adding it if new) and keeps at most max ids.
func addView(list []string, id string, max int) []string {
	out := []string{id}
	for _, v := range list {
		if v != id && len(out) < max {
			out = append(out, v)
		}
	}
	return out
}

// mergeViews puts the guest's views (newest first) in front of the account's
// own, without duplicates, keeping at most max ids. The guest's views are the
// newest because they happened just before signing in.
func mergeViews(guest, account []string, max int) []string {
	out := account
	for i := len(guest) - 1; i >= 0; i-- { // oldest first, so the newest ends up in front
		out = addView(out, guest[i], max)
	}
	return out
}

// recommend scores every product that is not in the history by what it shares
// with each history product: same genre +3, same category +2, same platform +1.
// Highest score first; ties (and users with no history) fall back to best sellers.
// Simple and explainable; a real shop would use co-purchase data instead.
func recommend(all, history []domain.Product, n int) []domain.Product {
	seen := map[string]bool{}
	for _, h := range history {
		seen[h.ID] = true
	}
	score := map[string]int{}
	var out []domain.Product
	for _, p := range all {
		if seen[p.ID] {
			continue
		}
		for _, h := range history {
			if p.Genre != "" && p.Genre == h.Genre {
				score[p.ID] += 3
			}
			if p.Category == h.Category {
				score[p.ID] += 2
			}
			if p.Platform == h.Platform {
				score[p.ID]++
			}
		}
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if score[out[i].ID] != score[out[j].ID] {
			return score[out[i].ID] > score[out[j].ID]
		}
		return out[i].Sales > out[j].Sales
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
