package service

import (
	"context"
	"errors"
	"slices"
	"sort"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
)

// maxViewed is how many recently viewed products a profile keeps.
const maxViewed = 20

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
	historyIDs := slices.Clone(viewed)
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
