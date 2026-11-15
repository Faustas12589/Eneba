import { getIdToken } from "./firebase";
import { guestId } from "./guest";
import type { Banner, CartUpdate, Order, OrderPage, Product, ProductDetail, ProductPage } from "../types";

const BASE = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");

/** Non-2xx response. `body` is the parsed JSON error (e.g. {stock, quantity} on a 409). */
export class ApiError extends Error {
  status: number;
  body: unknown;
  constructor(status: number, statusText: string, body: unknown) {
    super(`${status} ${statusText}`);
    this.status = status;
    this.body = body;
  }
}

interface RequestOpts {
  auth?: boolean;
  guest?: boolean; // send this browser's guest id (orders, while login is off)
  signal?: AbortSignal;
  method?: "GET" | "PUT" | "POST";
  body?: unknown; // sent as JSON
}

async function request<T>(path: string, opts: RequestOpts = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (opts.auth) {
    const token = await getIdToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }
  if (opts.guest) headers["X-Guest-Id"] = guestId();
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";
  const res = await fetch(`${BASE}${path}`, {
    method: opts.method ?? "GET",
    headers,
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
    signal: opts.signal,
  });
  if (!res.ok) throw new ApiError(res.status, res.statusText, await res.json().catch(() => null));
  return res.json() as Promise<T>;
}

export interface ProductQuery {
  q?: string;
  platform?: string[];
  region?: string[];
  genre?: string[];
  category?: string[];
  minPrice?: number; // cents
  maxPrice?: number; // cents
  sort?: string;
  page?: number;
}

export function toSearchParams(q: ProductQuery): URLSearchParams {
  const p = new URLSearchParams();
  if (q.q) p.set("q", q.q);
  if (q.platform?.length) p.set("platform", q.platform.join(","));
  if (q.region?.length) p.set("region", q.region.join(","));
  if (q.genre?.length) p.set("genre", q.genre.join(","));
  if (q.category?.length) p.set("category", q.category.join(","));
  if (q.minPrice) p.set("minPrice", String(q.minPrice));
  if (q.maxPrice) p.set("maxPrice", String(q.maxPrice));
  if (q.sort) p.set("sort", q.sort);
  if (q.page && q.page > 1) p.set("page", String(q.page));
  return p;
}

export const api = {
  products: (q: ProductQuery, signal?: AbortSignal) =>
    request<ProductPage>(`/api/products?${toSearchParams(q)}`, { signal }),
  product: (id: string) => request<ProductDetail>(`/api/products/${encodeURIComponent(id)}`),
  suggest: (q: string, signal?: AbortSignal) =>
    request<{ items: Product[] }>(`/api/search?q=${encodeURIComponent(q)}`, { signal }),
  banners: () => request<{ items: Banner[] }>("/api/home/banners"),
  bestSellers: () => request<{ items: Product[] }>("/api/home/best-sellers"),
  deals: () => request<{ items: Product[] }>("/api/home/deals"),
  /**
   * Reserve/release stock so the cart holds exactly `quantity` (0 = remove).
   * Returns the cart as stored. 409 = not enough stock, 410 = cart already paid.
   */
  setCartItem: (cartId: string, offerId: string, quantity: number) =>
    request<CartUpdate>(`/api/carts/${encodeURIComponent(cartId)}/items/${encodeURIComponent(offerId)}`, {
      method: "PUT",
      body: { quantity },
    }),
  /** The cart order as stored (expired lines are gone). 410 = cart already paid. */
  cart: (cartId: string) => request<Order>(`/api/carts/${encodeURIComponent(cartId)}`),
  me: () => request<{ uid: string; email: string }>("/api/me", { auth: true }),
  /** Simulated payment: turns the cart into a delivered order with its keys. Safe to retry. */
  checkout: (cartId: string) =>
    request<Order>("/api/orders", { guest: true, method: "POST", body: { cartId } }),
  /** One page of this browser's orders, newest first. */
  orders: (page: number, signal?: AbortSignal) =>
    request<OrderPage>(`/api/orders?page=${page}`, { guest: true, signal }),
};
