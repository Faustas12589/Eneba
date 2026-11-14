// Pure cart logic: no React, no storage, no network. Every function returns a new array.
import type { Offer, Product } from "../../types";

/** Same cap as the backend (service.MaxPerLine). */
export const MAX_PER_LINE = 10;

/**
 * One line per OFFER (a product can be in the cart from two sellers).
 * `quantity` is what the server has reserved for this cart; `stock` is how many
 * more units are still available to anyone (both come from the server's reply).
 * Name/price are a display snapshot; checkout re-reads prices on the server.
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

export function quantityOf(items: CartItem[], offerId: string): number {
  return items.find((i) => i.offerId === offerId)?.quantity ?? 0;
}

/** Saves the server's answer for one offer: new quantity (0 = line removed) and stock. */
export function applyServerLine(items: CartItem[], offerId: string, quantity: number, stock: number): CartItem[] {
  if (quantity === 0) return items.filter((i) => i.offerId !== offerId);
  return items.map((i) => (i.offerId === offerId ? { ...i, quantity, stock } : i));
}

/** Adds a line the server just reserved (or updates it if it's already there). */
export function addServerLine(items: CartItem[], product: Product, offer: Offer, quantity: number, stock: number): CartItem[] {
  if (items.some((i) => i.offerId === offer.id)) return applyServerLine(items, offer.id, quantity, stock);
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
      stock,
      quantity,
    },
  ];
}

/** Error text when the server refused because stock ran out. */
export function outOfStockMessage(stock: number, inCart: number): string {
  if (stock > 0) return `Liko tik ${stock} vnt.`;
  if (inCart > 0) return "Daugiau nebeliko: visos likusios prekės jau tavo krepšelyje.";
  return "Šis pasiūlymas jau išparduotas.";
}

export function itemCount(items: CartItem[]): number {
  return items.reduce((n, i) => n + i.quantity, 0);
}

export function totalCents(items: CartItem[]): number {
  return items.reduce((sum, i) => sum + i.price * i.quantity, 0);
}
