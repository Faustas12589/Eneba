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
	Genre           string `json:"genre" firestore:"genre"`
	Description     string `json:"description" firestore:"description"`
	ImageURL        string `json:"imageUrl" firestore:"imageUrl"`
	ListPrice       int    `json:"listPrice" firestore:"listPrice"` // cents, before discount
	PriceFrom       int    `json:"priceFrom" firestore:"priceFrom"` // cents, cheapest offer
	Currency        string `json:"currency" firestore:"currency"`
	DiscountPercent int    `json:"discountPercent" firestore:"discountPercent"`
	Sales           int    `json:"sales" firestore:"sales"`
}

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

// Reservation is how many units of one offer a cart is holding. Those units are
// already taken out of Offer.Stock, so Offer.Stock is what anyone else can still add.
type Reservation struct {
	CartID    string    `json:"cartId" firestore:"cartId"`
	OfferID   string    `json:"offerId" firestore:"offerId"`
	Quantity  int       `json:"quantity" firestore:"quantity"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// TODO (checkout slice): Order, OrderItem, DigitalKey.
