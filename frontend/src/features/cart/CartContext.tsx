import { createContext, useContext, useEffect, useState } from "react";
import type { Offer, Product } from "../../types";
import { addItem, itemCount, removeItem, setQuantity, totalCents, type CartItem } from "./cart";

interface CartValue {
  items: CartItem[];
  count: number; // total quantity, for the header badge
  total: number; // cents
  add: (product: Product, offer: Offer) => void;
  setQty: (offerId: string, qty: number) => void;
  remove: (offerId: string) => void;
  clear: () => void;
}

const Ctx = createContext<CartValue | null>(null);
const STORAGE_KEY = "cart";

function load(): CartItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    const parsed: unknown = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? (parsed as CartItem[]) : [];
  } catch {
    return []; // corrupt JSON or storage blocked: start empty
  }
}

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<CartItem[]>(load);

  // Persist on every change.
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch {
      // storage blocked: cart just won't survive a reload
    }
  }, [items]);

  const value: CartValue = {
    items,
    count: itemCount(items),
    total: totalCents(items),
    add: (product, offer) => setItems((cur) => addItem(cur, product, offer)),
    setQty: (offerId, qty) => setItems((cur) => setQuantity(cur, offerId, qty)),
    remove: (offerId) => setItems((cur) => removeItem(cur, offerId)),
    clear: () => setItems([]),
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useCart(): CartValue {
  const v = useContext(Ctx);
  if (!v) throw new Error("useCart must be used inside <CartProvider>");
  return v;
}
