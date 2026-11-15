// Pure cart helpers: no React, no storage, no network.
// The cart itself is the server's order document (status "cart"); the browser
// never builds or stores its own copy of the lines.
import type { Order, OrderItem } from "../../types";

/** Same cap as the backend (service.MaxPerLine). */
export const MAX_PER_LINE = 10;

/** An empty cart, shown before the server answered or after checkout. */
export function emptyCart(id: string): Order {
  return { id, status: "cart", items: [], total: 0, currency: "EUR", createdAt: "", paidAt: null, deliveredAt: null };
}

export function quantityOf(items: OrderItem[], offerId: string): number {
  return items.find((i) => i.offerId === offerId)?.quantity ?? 0;
}

/** True if the server holds less than before: a line expired or shrank. */
export function lostLines(before: OrderItem[], after: OrderItem[]): boolean {
  return before.some((b) => quantityOf(after, b.offerId) < b.quantity);
}

/** Error text when the server refused because stock ran out. */
export function outOfStockMessage(stock: number, inCart: number): string {
  if (stock > 0) return `Liko tik ${stock} vnt.`;
  if (inCart > 0) return "Daugiau nebeliko: visos likusios prekės jau tavo krepšelyje.";
  return "Šis pasiūlymas jau išparduotas.";
}

export function itemCount(items: OrderItem[]): number {
  return items.reduce((n, i) => n + i.quantity, 0);
}
