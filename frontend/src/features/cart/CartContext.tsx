import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import { api } from "../../lib/api";
import type { Offer, Product } from "../../types";
import {
  addError,
  addItem,
  itemCount,
  quantityOf,
  removeItem,
  setQuantity,
  totalCents,
  updateStock,
  type CartItem,
} from "./cart";

/** What the add-to-cart popup shows: success, or why the add failed. */
export type CartNotice = { kind: "added" } | { kind: "error"; message: string };

interface CartValue {
  items: CartItem[];
  count: number; // total quantity, for the header badge and popup
  total: number; // cents
  /** Checks live stock on the server, then adds one unit and opens the popup. */
  add: (product: Product, offer: Offer) => Promise<void>;
  /** Checks live stock, then adds one unit. Returns an error message or null. */
  increase: (offerId: string) => Promise<string | null>;
  setQty: (offerId: string, qty: number) => void;
  remove: (offerId: string) => void;
  clear: () => void;
  notice: CartNotice | null;
  closeNotice: () => void;
}

const Ctx = createContext<CartValue | null>(null);
const STORAGE_KEY = "cart";
const CHECK_FAILED = "Nepavyko patikrinti likučio. Bandykite dar kartą.";

function load(): CartItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    const parsed: unknown = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? (parsed as CartItem[]) : [];
  } catch {
    return []; // corrupt JSON or storage blocked: start empty
  }
}

/** Current stock of one offer, straight from the server (0 if the offer is gone). */
async function fetchStock(productId: string, offerId: string): Promise<number> {
  const { offers } = await api.product(productId);
  return offers.find((o) => o.id === offerId)?.stock ?? 0;
}

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<CartItem[]>(load);
  const [notice, setNotice] = useState<CartNotice | null>(null);
  // Latest items for async functions: after `await`, the `items` variable
  // they closed over may be out of date.
  const itemsRef = useRef(items);

  useEffect(() => {
    itemsRef.current = items;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch {
      // storage blocked: cart just won't survive a reload
    }
  }, [items]);

  const add = async (product: Product, offer: Offer) => {
    let stock: number;
    try {
      stock = await fetchStock(product.id, offer.id);
    } catch {
      setNotice({ kind: "error", message: CHECK_FAILED });
      return;
    }
    const error = addError(itemsRef.current, offer.id, stock);
    if (error) {
      setItems((cur) => updateStock(cur, offer.id, stock));
      setNotice({ kind: "error", message: error });
      return;
    }
    setItems((cur) => addItem(cur, product, { ...offer, stock }));
    setNotice({ kind: "added" });
  };

  const increase = async (offerId: string): Promise<string | null> => {
    const item = itemsRef.current.find((i) => i.offerId === offerId);
    if (!item) return null;
    let stock: number;
    try {
      stock = await fetchStock(item.productId, offerId);
    } catch {
      return CHECK_FAILED;
    }
    setItems((cur) => updateStock(cur, offerId, stock));
    const error = addError(itemsRef.current, offerId, stock);
    if (error) return error;
    setItems((cur) => setQuantity(cur, offerId, quantityOf(cur, offerId) + 1));
    return null;
  };

  const closeNotice = useCallback(() => setNotice(null), []);

  const value: CartValue = {
    items,
    count: itemCount(items),
    total: totalCents(items),
    add,
    increase,
    setQty: (offerId, qty) => setItems((cur) => setQuantity(cur, offerId, qty)),
    remove: (offerId) => setItems((cur) => removeItem(cur, offerId)),
    clear: () => setItems([]),
    notice,
    closeNotice,
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useCart(): CartValue {
  const v = useContext(Ctx);
  if (!v) throw new Error("useCart must be used inside <CartProvider>");
  return v;
}
