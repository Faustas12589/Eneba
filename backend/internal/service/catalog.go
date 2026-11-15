// Package service contains business logic. It knows nothing about HTTP.
package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
	"eneba-marketplace/backend/internal/search"
)

const cacheTTL = 60 * time.Second

type Catalog struct {
	repo repository.Catalog

	mu       sync.Mutex
	loadedAt time.Time
	products []domain.Product
	byID     map[string]domain.Product
	index    *search.Index
}

func NewCatalog(repo repository.Catalog) *Catalog { return &Catalog{repo: repo} }

// snapshot returns the cached product list, reloading when stale.
// Filtering/sorting happens in Go because Firestore can't combine arbitrary
// filters + sorts without a composite index per combination (see DECISIONS.md).
func (c *Catalog) snapshot(ctx context.Context) ([]domain.Product, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.products != nil && time.Since(c.loadedAt) < cacheTTL {
		return c.products, nil
	}
	ps, err := c.repo.ListProducts(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]search.Item, len(ps))
	byID := make(map[string]domain.Product, len(ps))
	for i := range ps {
		ps[i].DiscountPercent = discountPercent(ps[i].ListPrice, ps[i].PriceFrom)
		p := ps[i]
		items[i] = search.Item{ID: p.ID, Name: p.Name}
		byID[p.ID] = p
	}
	c.products, c.byID, c.index, c.loadedAt = ps, byID, search.Build(items), time.Now()
	return ps, nil
}

// discountPercent is derived from prices instead of trusting a stored value,
// so the badge can never disagree with the prices shown. Rounded to nearest %.
// 0 when the cheapest offer is not below the list price.
func discountPercent(listPrice, priceFrom int) int {
	if listPrice <= 0 || priceFrom >= listPrice {
		return 0
	}
	return ((listPrice-priceFrom)*100 + listPrice/2) / listPrice
}

type Filter struct {
	Q          string
	Platforms  []string
	Regions    []string
	Genres     []string
	Categories []string
	MinPrice   int // cents, 0 = unset
	MaxPrice   int // cents, 0 = unset
	Sort       string
	Page       int
	PageSize   int
}

type Facets struct {
	Platforms  []string `json:"platforms"`
	Regions    []string `json:"regions"`
	Genres     []string `json:"genres"`
	Categories []string `json:"categories"`
}

type Page struct {
	Items    []domain.Product `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Facets   Facets           `json:"facets"`
}

func (c *Catalog) List(ctx context.Context, f Filter) (Page, error) {
	all, err := c.snapshot(ctx)
	if err != nil {
		return Page{}, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 60 {
		f.PageSize = 24
	}

	// Text search first: gives relevance ordering.
	rank := map[string]int{}
	pool := all
	if strings.TrimSpace(f.Q) != "" {
		c.mu.Lock()
		res := c.index.Query(f.Q, 0)
		c.mu.Unlock()
		pool = pool[:0:0]
		for i, r := range res {
			rank[r.ID] = i
			pool = append(pool, c.byIDLocked(r.ID))
		}
	}

	out := make([]domain.Product, 0, len(pool))
	for _, p := range pool {
		if !inSet(f.Platforms, p.Platform) || !inSet(f.Regions, p.Region) || !inSet(f.Genres, p.Genre) ||
			!inSet(f.Categories, p.Category) {
			continue
		}
		if f.MinPrice > 0 && p.PriceFrom < f.MinPrice {
			continue
		}
		if f.MaxPrice > 0 && p.PriceFrom > f.MaxPrice {
			continue
		}
		out = append(out, p)
	}

	sortProducts(out, f.Sort, rank, f.Q != "")

	total := len(out)
	start, end := pageBounds(f.Page, f.PageSize, total)
	return Page{Items: out[start:end], Total: total, Page: f.Page, PageSize: f.PageSize, Facets: facets(all)}, nil
}

// pageBounds returns the slice range [start, end) of a 1-based page. A page
// past the end gives an empty range instead of an out-of-bounds slice.
func pageBounds(page, pageSize, total int) (start, end int) {
	start = min((page-1)*pageSize, total)
	end = min(start+pageSize, total)
	return start, end
}

func (c *Catalog) byIDLocked(id string) domain.Product {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.byID[id]
}

func sortProducts(ps []domain.Product, sortBy string, rank map[string]int, hasQuery bool) {
	less := map[string]func(a, b domain.Product) bool{
		"price_asc":  func(a, b domain.Product) bool { return a.PriceFrom < b.PriceFrom },
		"price_desc": func(a, b domain.Product) bool { return a.PriceFrom > b.PriceFrom },
		"discount":   func(a, b domain.Product) bool { return a.DiscountPercent > b.DiscountPercent },
		"name":       func(a, b domain.Product) bool { return a.Name < b.Name },
		"popular":    func(a, b domain.Product) bool { return a.Sales > b.Sales },
	}
	if sortBy == "" || sortBy == "relevance" {
		if hasQuery {
			sort.SliceStable(ps, func(i, j int) bool { return rank[ps[i].ID] < rank[ps[j].ID] })
			return
		}
		sortBy = "popular"
	}
	fn, ok := less[sortBy]
	if !ok {
		fn = less["popular"]
	}
	sort.SliceStable(ps, func(i, j int) bool { return fn(ps[i], ps[j]) })
}

func inSet(set []string, v string) bool {
	if len(set) == 0 {
		return true
	}
	for _, s := range set {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}

func facets(ps []domain.Product) Facets {
	pl, rg, gn, ct := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, p := range ps {
		pl[p.Platform], rg[p.Region], gn[p.Genre], ct[p.Category] = true, true, true, true
	}
	// Categories keep their fixed menu order; the rest are alphabetical.
	cats := []string{}
	for _, c := range domain.Categories {
		if ct[c] {
			cats = append(cats, c)
		}
	}
	return Facets{Platforms: keys(pl), Regions: keys(rg), Genres: keys(gn), Categories: cats}
}

// keys returns the map's keys sorted, without "" (e.g. cards have no genre).
func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		if k != "" {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

type ProductDetail struct {
	Product domain.Product `json:"product"`
	Offers  []domain.Offer `json:"offers"` // cheapest first
}

func (c *Catalog) Get(ctx context.Context, id string) (*ProductDetail, error) {
	p, err := c.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	p.DiscountPercent = discountPercent(p.ListPrice, p.PriceFrom)
	offers, err := c.repo.ListOffers(ctx, id)
	if err != nil {
		return nil, err
	}
	sort.Slice(offers, func(i, j int) bool { return offers[i].Price < offers[j].Price })
	return &ProductDetail{Product: *p, Offers: offers}, nil
}

// Suggest powers the header search box dropdown.
func (c *Catalog) Suggest(ctx context.Context, q string, limit int) ([]domain.Product, error) {
	pg, err := c.List(ctx, Filter{Q: q, PageSize: limit})
	return pg.Items, err
}

// BestSellers and Deals are separate "services" so the homepage blocks are
// independent: each has its own endpoint and can fail on its own.
func (c *Catalog) BestSellers(ctx context.Context, n int) ([]domain.Product, error) {
	pg, err := c.List(ctx, Filter{Sort: "popular", PageSize: n})
	return pg.Items, err
}

func (c *Catalog) Deals(ctx context.Context, n int) ([]domain.Product, error) {
	pg, err := c.List(ctx, Filter{Sort: "discount", PageSize: n})
	return pg.Items, err
}

// lookup returns the products with these ids, in the same order. Unknown ids
// (e.g. a product removed since someone viewed it) are skipped.
func (c *Catalog) lookup(ctx context.Context, ids []string) ([]domain.Product, error) {
	if _, err := c.snapshot(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := c.byID[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (c *Catalog) Banners(ctx context.Context) ([]domain.Banner, error) {
	return c.repo.ListBanners(ctx)
}
