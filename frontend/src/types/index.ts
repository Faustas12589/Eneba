// Mirrors backend/internal/domain. Money is integer cents.
export interface Product {
  id: string;
  name: string;
  slug: string;
  platform: string;
  region: string;
  genre: string; // "" for cards and software
  category: string; // id from lib/categories.ts
  description: string;
  imageUrl: string;
  listPrice: number;
  priceFrom: number;
  currency: string;
  discountPercent: number;
  sales: number;
  releaseDate?: string; // "YYYY-MM-DD", only on unreleased products
  preOrder: boolean; // derived by the server from releaseDate
}

/** GET /api/home/categories: products per category, in menu order. */
export interface CategoryCount {
  category: string;
  count: number;
}

export interface Offer {
  id: string;
  productId: string;
  sellerId: string;
  sellerName: string;
  sellerRating: number;
  price: number;
  currency: string;
  stock: number;
}

export interface Banner {
  id: string;
  title: string;
  subtitle: string;
  imageUrl: string;
  linkUrl: string;
}

export interface Facets {
  platforms: string[];
  regions: string[];
  genres: string[];
  categories: string[];
  priceMin: number; // cents: cheapest product matching every filter except price
  priceMax: number; // cents
}

export interface ProductPage {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
  facets: Facets;
}

export interface ProductDetail {
  product: Product;
  offers: Offer[];
}

/** cart -> paid -> delivered. The cart is the first stage of an order. */
export type OrderStatus = "cart" | "paid" | "delivered";

/** Mirrors domain.OrderItem: written whole when the line goes into the cart. */
export interface OrderItem {
  offerId: string;
  productId: string;
  productName: string;
  imageUrl: string;
  sellerName: string;
  price: number; // cents per unit, as when the line was added
  quantity: number;
  expiresAt: string | null; // reservation end while in the cart; null once paid
  keys: { code: string }[] | null; // the server only shows codes once delivered
}

/** Mirrors domain.Order. A cart is an order with status "cart". */
export interface Order {
  id: string;
  status: OrderStatus;
  items: OrderItem[];
  total: number; // cents
  currency: string;
  createdAt: string;
  paidAt: string | null;
  deliveredAt: string | null;
}

/** GET /api/orders: one page of order history, newest first. */
export interface OrderPage {
  items: Order[];
  total: number;
  page: number;
  pageSize: number;
}

/** PUT /api/carts/{id}/items/{offerId}: the cart as stored + units left of that offer. */
export interface CartUpdate {
  cart: Order;
  offerId: string;
  stock: number;
}

/** Body of a 409 on a cart change: units left, and what this cart still holds. */
export interface OutOfStock {
  stock: number;
  quantity: number;
}

/** GET /api/me: purchase numbers for the account page (name, photo etc. come from Firebase Auth). */
export interface Account {
  uid: string;
  email: string;
  orderCount: number;
  totalSpent: number; // cents
}
