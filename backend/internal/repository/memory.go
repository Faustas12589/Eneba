package repository

import (
	"context"
	"slices"
	"sync"
	"time"

	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/seeds"
)

// Memory serves the embedded seed data. Handy for local dev and tests
// before Firebase credentials are set up. Changes live until restart.
type Memory struct {
	products []domain.Product
	banners  []domain.Banner

	mu       sync.Mutex // guards offers (stock), keys, orders and profiles
	offers   []domain.Offer
	keys     []domain.Key
	orders   map[string]domain.Order
	profiles map[string]domain.Profile
}

func NewMemory() (*Memory, error) {
	m := &Memory{orders: map[string]domain.Order{}, profiles: map[string]domain.Profile{}}
	if err := seeds.Load("products", &m.products); err != nil {
		return nil, err
	}
	if err := seeds.Load("offers", &m.offers); err != nil {
		return nil, err
	}
	if err := seeds.Load("banners", &m.banners); err != nil {
		return nil, err
	}
	m.keys = seeds.Keys(m.offers)
	return m, nil
}

func (m *Memory) ListProducts(context.Context) ([]domain.Product, error) {
	return append([]domain.Product(nil), m.products...), nil
}

func (m *Memory) GetProduct(_ context.Context, id string) (*domain.Product, error) {
	for i := range m.products {
		if m.products[i].ID == id {
			p := m.products[i]
			return &p, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) ListOffers(_ context.Context, productID string) ([]domain.Offer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Offer
	for _, o := range m.offers {
		if o.ProductID == productID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (m *Memory) ListBanners(context.Context) ([]domain.Banner, error) {
	return append([]domain.Banner(nil), m.banners...), nil
}

// UpdateReservation holds the mutex for the whole read-apply-write, which is
// the in-memory equivalent of a Firestore transaction.
func (m *Memory) UpdateReservation(ctx context.Context, cartID, offerID string, apply ApplyFunc) (*domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	offer := m.offer(offerID)
	if offer == nil {
		return nil, ErrNotFound
	}
	cart, ok := m.orders[cartID]
	if ok && cart.Status != domain.OrderCart {
		return nil, ErrCartClosed
	}
	if !ok {
		cart = domain.Order{ID: cartID, Status: domain.OrderCart, Currency: offer.Currency, CreatedAt: time.Now()}
	}
	cart = cloneOrder(cart)

	cur := cart.Item(offerID)
	before := reservationOf(cartID, cur)
	newStock, next, err := apply(offer.Stock, before)
	if err != nil {
		return nil, err
	}
	need := next.Quantity - before.Quantity // keys to reserve; < 0 means some go back

	var product domain.Product
	if cur == nil && next.Quantity > 0 {
		p, err := m.GetProduct(ctx, offer.ProductID)
		if err != nil {
			return nil, err
		}
		product = *p
	}
	// Pick keys first, change anything only once we know there are enough.
	var picked []int
	for i, k := range m.keys {
		if len(picked) < need && k.OfferID == offerID && k.Status == domain.KeyAvailable {
			picked = append(picked, i)
		}
	}
	if len(picked) < need {
		return nil, errKeyPoolMismatch
	}
	added := make([]domain.Key, len(picked))
	for j, i := range picked {
		m.keys[i].Status, m.keys[i].OrderID = domain.KeyReserved, cartID
		added[j] = m.keys[i]
	}

	item, released := nextItem(cur, *offer, product, next, added)
	for _, rk := range released {
		if k := m.key(rk.ID); k != nil {
			k.Status, k.OrderID = domain.KeyAvailable, ""
		}
	}
	offer.Stock = newStock
	cart.PutItem(item)
	if len(cart.Items) == 0 {
		delete(m.orders, cartID)
	} else {
		m.orders[cartID] = cart
	}
	out := cloneOrder(cart)
	return &out, nil
}

func (m *Memory) ListExpiredReservations(_ context.Context, now time.Time) ([]domain.Reservation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Reservation
	for _, o := range m.orders {
		out = append(out, expiredLines(o, now)...)
	}
	return out, nil
}

func (m *Memory) GetOrder(_ context.Context, id string) (*domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	o = cloneOrder(o)
	return &o, nil
}

func (m *Memory) PayOrder(_ context.Context, id string, apply PayFunc) (*domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	o = cloneOrder(o)
	if err := apply(&o); err != nil {
		return nil, err
	}
	m.orders[id] = o
	out := cloneOrder(o)
	return &out, nil
}

func (m *Memory) DeliverOrder(_ context.Context, id string, apply DeliverFunc) (*domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	o = cloneOrder(o)
	var keys []domain.Key
	for _, it := range o.Items {
		for _, ik := range it.Keys {
			if k := m.key(ik.ID); k != nil {
				keys = append(keys, *k)
			}
		}
	}
	if err := apply(&o, keys); err != nil {
		return nil, err
	}
	for _, k := range keys {
		*m.key(k.ID) = k
	}
	m.orders[id] = o
	out := cloneOrder(o)
	return &out, nil
}

func (m *Memory) ListUserOrders(_ context.Context, userID string) ([]domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Order
	for _, o := range m.orders {
		if o.UserID == userID {
			out = append(out, cloneOrder(o))
		}
	}
	return out, nil
}

func (m *Memory) MoveUserOrders(_ context.Context, from, to string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for id, o := range m.orders {
		if o.UserID == from {
			o.UserID = to
			m.orders[id] = o
			n++
		}
	}
	return n, nil
}

// KeyCounts returns how many of an offer's keys are in each status. Used by
// tests to check that offer stock always equals the available keys.
func (m *Memory) KeyCounts(offerID string) map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	counts := map[string]int{}
	for _, k := range m.keys {
		if k.OfferID == offerID {
			counts[k.Status]++
		}
	}
	return counts
}

// offer and key return pointers into the slices (caller holds the mutex), or nil.
func (m *Memory) offer(id string) *domain.Offer {
	for i := range m.offers {
		if m.offers[i].ID == id {
			return &m.offers[i]
		}
	}
	return nil
}

func (m *Memory) key(id string) *domain.Key {
	for i := range m.keys {
		if m.keys[i].ID == id {
			return &m.keys[i]
		}
	}
	return nil
}

// cloneOrder copies the slices too, so callers can't change stored orders.
func cloneOrder(o domain.Order) domain.Order {
	o.Items = slices.Clone(o.Items)
	for i := range o.Items {
		o.Items[i].Keys = slices.Clone(o.Items[i].Keys)
	}
	return o
}

func (m *Memory) GetProfile(_ context.Context, userID string) (*domain.Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[userID]
	if !ok {
		return nil, ErrNotFound
	}
	p.RecentlyViewed = slices.Clone(p.RecentlyViewed) // callers can't change the stored copy
	return &p, nil
}

func (m *Memory) UpdateProfile(_ context.Context, userID string, apply ProfileFunc) (*domain.Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[userID]
	if !ok {
		p = domain.Profile{UserID: userID}
	}
	p.RecentlyViewed = slices.Clone(p.RecentlyViewed)
	if err := apply(&p); err != nil {
		return nil, err
	}
	m.profiles[userID] = p
	out := p
	out.RecentlyViewed = slices.Clone(p.RecentlyViewed)
	return &out, nil
}
