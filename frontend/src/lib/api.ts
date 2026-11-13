import { getIdToken } from "./firebase";
import type { Banner, Product, ProductDetail, ProductPage } from "../types";

const BASE = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");

async function request<T>(path: string, opts: { auth?: boolean; signal?: AbortSignal } = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (opts.auth) {
    const token = await getIdToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }
  const res = await fetch(`${BASE}${path}`, { headers, signal: opts.signal });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json() as Promise<T>;
}

export interface ProductQuery {
  q?: string;
  platform?: string[];
  region?: string[];
  genre?: string[];
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
  me: () => request<{ uid: string; email: string }>("/api/me", { auth: true }),
};
