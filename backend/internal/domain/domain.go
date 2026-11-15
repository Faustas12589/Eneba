// Package domain holds the plain data types shared by every layer.
// All money is stored as integer cents to avoid float rounding bugs.
package domain

import "time"

type Product struct {
	ID              string `json:"id" firestore:"id"`
	Name            string `json:"name" firestore:"name"`
	Slug            string `json:"slug" firestore:"slug"`
	Platform        string `json:"platform" firestore:"platform"`
	Region          string `json:"region" firestore:"region"`
	Genre           string `json:"genre" firestore:"genre"`       // games and DLC only; "" for cards and software
	Category        string `json:"category" firestore:"category"` // one of Categories
	Description     string `json:"description" firestore:"description"`
	ImageURL        string `json:"imageUrl" firestore:"imageUrl"`
	ListPrice       int    `json:"listPrice" firestore:"listPrice"` // cents, before discount
	PriceFrom       int    `json:"priceFrom" firestore:"priceFrom"` // cents, cheapest offer
	Currency        string `json:"currency" firestore:"currency"`
	DiscountPercent int    `json:"discountPercent" firestore:"discountPercent"`
	Sales           int    `json:"sales" firestore:"sales"`
	// ReleaseDate is "YYYY-MM-DD", or "" for products that are already out.
	ReleaseDate string `json:"releaseDate,omitempty" firestore:"releaseDate"`
	// PreOrder is derived from ReleaseDate when the catalog loads (never stored),
	// so a product stops being a pre-order by itself on release day.
	PreOrder bool `json:"preOrder" firestore:"-"`
}

// Categories are the product types, like Eneba's top-level menu. A product has
// exactly one. The order here is the order the store's filter shows them in;
// the frontend turns the ids into Lithuanian labels.
var Categories = []string{"games", "dlc", "gaming-cards", "gift-cards", "software"}

type Offer struct {
	ID           string  `json:"id" firestore:"id"`
	ProductID    string  `json:"productId" firestore:"productId"`
	SellerID     string  `json:"sellerId" firestore:"sellerId"`
	SellerName   string  `json:"sellerName" firestore:"sellerName"`
	SellerRating float64 `json:"sellerRating" firestore:"sellerRating"`
	Price        int     `json:"price" firestore:"price"` // cents
	Currency     string  `json:"currency" firestore:"currency"`
	Stock        int     `json:"stock" firestore:"stock"`
}

type Banner struct {
	ID       string `json:"id" firestore:"id"`
	Title    string `json:"title" firestore:"title"`
	Subtitle string `json:"subtitle" firestore:"subtitle"`
	ImageURL string `json:"imageUrl" firestore:"imageUrl"`
	LinkURL  string `json:"linkUrl" firestore:"linkUrl"`
}

// Reservation is one cart line seen from the stock side: how many units of an
// offer a cart holds and until when. It is not stored on its own: it is read
// from the cart's Order (status "cart"). Those units' keys are "reserved", so
// Offer.Stock (= available keys) is what anyone else can still add.
type Reservation struct {
	CartID    string
	OfferID   string
	Quantity  int
	ExpiresAt time.Time
}

// Key statuses: available -> reserved (in a cart) -> sold (delivered).
// A reserved key goes back to available if its cart line is removed or expires.
const (
	KeyAvailable = "available"
	KeyReserved  = "reserved"
	KeySold      = "sold"
)

// Key is one digital game key. Each offer has a pool of them; the offer's
// Stock always equals how many of its keys are available.
type Key struct {
	ID      string `json:"id" firestore:"id"`
	OfferID string `json:"offerId" firestore:"offerId"`
	Code    string `json:"code" firestore:"code"`
	Status  string `json:"status" firestore:"status"`
	OrderID string `json:"orderId" firestore:"orderId"` // empty while available
}

// Order statuses. The cart is the first stage of an order:
// cart (keys reserved) -> paid (owner fixed) -> delivered (keys sold, codes shown).
const (
	OrderCart      = "cart"
	OrderPaid      = "paid"
	OrderDelivered = "delivered"
)

// Order is a cart that becomes a purchase. Its ID is the browser's cart id.
// Everything about a line (name, price, keys) is written when the line goes
// into the cart; later stages only change status fields, never rebuild data.
type Order struct {
	ID          string      `json:"id" firestore:"id"`
	UserID      string      `json:"-" firestore:"userId"` // empty while it is a guest cart
	Status      string      `json:"status" firestore:"status"`
	Items       []OrderItem `json:"items" firestore:"items"`
	Total       int         `json:"total" firestore:"total"` // cents, kept up to date by every cart change
	Currency    string      `json:"currency" firestore:"currency"`
	CreatedAt   time.Time   `json:"createdAt" firestore:"createdAt"`
	PaidAt      *time.Time  `json:"paidAt" firestore:"paidAt"`           // null until paid
	DeliveredAt *time.Time  `json:"deliveredAt" firestore:"deliveredAt"` // null until delivered
}

// OrderItem is one offer in an order, with the keys reserved for it.
// Name, image, seller and price are a snapshot from when it was first added.
type OrderItem struct {
	OfferID     string     `json:"offerId" firestore:"offerId"`
	ProductID   string     `json:"productId" firestore:"productId"`
	ProductName string     `json:"productName" firestore:"productName"`
	ImageURL    string     `json:"imageUrl" firestore:"imageUrl"`
	SellerName  string     `json:"sellerName" firestore:"sellerName"`
	Price       int        `json:"price" firestore:"price"` // cents per unit
	Quantity    int        `json:"quantity" firestore:"quantity"`
	ExpiresAt   *time.Time `json:"expiresAt" firestore:"expiresAt"` // reservation end; null once paid
	Keys        []ItemKey  `json:"keys" firestore:"keys"`           // len == Quantity; codes hidden until delivered (Public)
}

// ItemKey is a key reserved for (later: sold with) an order line.
type ItemKey struct {
	ID   string `json:"-" firestore:"id"`
	Code string `json:"code" firestore:"code"`
}

// Item returns the line for an offer, or nil if the order has none.
func (o *Order) Item(offerID string) *OrderItem {
	for i := range o.Items {
		if o.Items[i].OfferID == offerID {
			return &o.Items[i]
		}
	}
	return nil
}

// PutItem replaces the line for it.OfferID (or adds it; Quantity 0 removes it)
// and updates the total.
func (o *Order) PutItem(it OrderItem) {
	var items []OrderItem
	replaced := false
	for _, x := range o.Items {
		if x.OfferID != it.OfferID {
			items = append(items, x)
		} else if it.Quantity > 0 {
			items = append(items, it) // keep the line's place in the list
			replaced = true
		}
	}
	if !replaced && it.Quantity > 0 {
		items = append(items, it)
	}
	o.Items = items
	o.Total = 0
	for _, x := range o.Items {
		o.Total += x.Price * x.Quantity
	}
}

// Public is the order as the API shows it: key codes are hidden until the
// order is delivered, so nobody can read keys from an unpaid cart.
func (o Order) Public() Order {
	if o.Status == OrderDelivered {
		return o
	}
	items := make([]OrderItem, len(o.Items))
	for i, it := range o.Items {
		it.Keys = nil
		items[i] = it
	}
	o.Items = items
	return o
}

// Profile is what the shop remembers about one user, so the homepage can show
// "recently viewed" and "recommended for you". UserID is whatever the request's
// user is: "guest:<id>" while login is off (DECISIONS #20), the Firebase uid later.
// Purchases are not copied here; they are read from the user's orders.
type Profile struct {
	UserID         string    `json:"userId" firestore:"userId"`
	RecentlyViewed []string  `json:"recentlyViewed" firestore:"recentlyViewed"` // product ids, newest first
	UpdatedAt      time.Time `json:"updatedAt" firestore:"updatedAt"`
}
