// Pure cart logic: no React, no storage. Every function returns a new array.
import type { Offer, Product } from "../../types";

/**
 * One line per OFFER (a product can be in the cart from two sellers).
 * Name/price/stock are a snapshot for display only; checkout re-reads them
 * on the server, so a stale or tampered value can't change what is charged.
 */
export interface CartItem {
  offerId: string;
  productId: string;
  productName: string;
  imageUrl: string;
  sellerName: string;
  price: number; // cents
  currency: string;
  stock: number;
  quantity: number;
}

/** Keeps quantity between 1 and the offer's stock. */
function clampQty(qty: number, stock: number): number {
  return Math.max(1, Math.min(Math.floor(qty), stock));
}

export function quantityOf(items: CartItem[], offerId: string): number {
  return items.find((i) => i.offerId === offerId)?.quantity ?? 0;
}

/**
 * Why one more unit of `offerId` can't be added when the server says `stock`
 * are left, or null if it can. Covers both "you already have all of it" and
 * "someone else bought it since you loaded the page" (stock went down).
 */
export function addError(items: CartItem[], offerId: string, stock: number): string | null {
  if (stock < 1) return "Šis pasiūlymas jau išparduotas.";
  if (quantityOf(items, offerId) >= stock) {
    return `Daugiau pridėti negalima: liko tik ${stock} vnt., ir visi jie jau tavo krepšelyje.`;
  }
  return null;
}

/** Replaces the stock snapshot with the server's current value. */
export function updateStock(items: CartItem[], offerId: string, stock: number): CartItem[] {
  return items.map((i) => (i.offerId === offerId ? { ...i, stock } : i));
}

/** Adds one unit. `offer.stock` should be fresh from the server. */
export function addItem(items: CartItem[], product: Product, offer: Offer): CartItem[] {
  if (offer.stock < 1) return items;
  const existing = items.find((i) => i.offerId === offer.id);
  if (existing) return setQuantity(updateStock(items, offer.id, offer.stock), offer.id, existing.quantity + 1);
  return [
    ...items,
    {
      offerId: offer.id,
      productId: product.id,
      productName: product.name,
      imageUrl: product.imageUrl,
      sellerName: offer.sellerName,
      price: offer.price,
      currency: offer.currency,
      stock: offer.stock,
      quantity: 1,
    },
  ];
}

export function setQuantity(items: CartItem[], offerId: string, qty: number): CartItem[] {
  if (!Number.isFinite(qty)) return items;
  return items.map((i) => (i.offerId === offerId ? { ...i, quantity: clampQty(qty, i.stock) } : i));
}

export function removeItem(items: CartItem[], offerId: string): CartItem[] {
  return items.filter((i) => i.offerId !== offerId);
}

export function itemCount(items: CartItem[]): number {
  return items.reduce((n, i) => n + i.quantity, 0);
}

export function totalCents(items: CartItem[]): number {
  return items.reduce((sum, i) => sum + i.price * i.quantity, 0);
}
