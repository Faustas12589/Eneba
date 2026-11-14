// Mirrors backend/internal/domain. Money is integer cents.
export interface Product {
  id: string;
  name: string;
  slug: string;
  platform: string;
  region: string;
  genre: string;
  description: string;
  imageUrl: string;
  listPrice: number;
  priceFrom: number;
  currency: string;
  discountPercent: number;
  sales: number;
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

/** Server's answer to a cart change (and the body of a 409, minus offerId). */
export interface CartLine {
  offerId: string;
  quantity: number; // units this cart now reserves
  stock: number; // units still available to everyone else
}
