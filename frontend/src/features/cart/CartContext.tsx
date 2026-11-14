import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import { api, ApiError } from "../../lib/api";
import type { Offer, Product } from "../../types";
import {
  addServerLine,
  applyServerLine,
  itemCount,
  MAX_PER_LINE,
  outOfStockMessage,
  quantityOf,
  reconcile,
  totalCents,
  type CartItem,
} from "./cart";

/** What the add-to-cart popup shows: success, or why the add failed. */
export type CartNotice = { kind: "added" } | { kind: "error"; title?: string; message: string };

interface CartValue {
  items: CartItem[];
  count: number; // total quantity, for the header badge and popup
  total: number; // cents
  /** Latest stock the server reported per offer id (for the product page table). */
  liveStock: Record<string, number>;
  /** Reserves one more unit on the server, then opens the popup (success or error). */
  add: (product: Product, offer: Offer) => Promise<void>;
  /** Sets a line's quantity on the server (0 = remove). Returns an error message or null. */
  setQty: (offerId: string, qty: number) => Promise<string | null>;
  notice: CartNotice | null;
  closeNotice: () => void;
}

const Ctx = createContext<CartValue | null>(null);
// v2: lines are now server reservations; older local-only carts are dropped.
const ITEMS_KEY = "cart.v2";
const ID_KEY = "cartId";
const FAILED = "Nepavyko atnaujinti krepšelio. Bandykite dar kartą.";
const EXPIRED = "Rezervacija baigėsi: kai kurios prekės grąžintos į sandėlį ir pašalintos iš krepšelio.";
const TOO_MANY = `Vienam pasiūlymui galima daugiausia ${MAX_PER_LINE} vnt.`;

function loadItems(): CartItem[] {
  try {
    const raw = localStorage.getItem(ITEMS_KEY);
    const parsed: unknown = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? (parsed as CartItem[]) : [];
  } catch {
    return []; // corrupt JSON or storage blocked: start empty
  }
}

/** Random id that names this browser's cart on the server (like a guest session). */
function loadCartId(): string {
  try {
    let id = localStorage.getItem(ID_KEY);
    if (!id) {
      id = crypto.randomUUID();
      localStorage.setItem(ID_KEY, id);
    }
    return id;
  } catch {
    return crypto.randomUUID(); // storage blocked: cart lives for this page load only
  }
}

type SyncResult =
  | { ok: true; quantity: number; stock: number; expiresAt?: string }
  | { ok: false; error: string };

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<CartItem[]>(loadItems);
  const [cartId] = useState(loadCartId);
  const [liveStock, setLiveStock] = useState<Record<string, number>>({});
  const [notice, setNotice] = useState<CartNotice | null>(null);
  // Latest items for async functions: after `await`, the `items` variable
  // they closed over may be out of date.
  const itemsRef = useRef(items);
  // Changes in flight + a counter bumped by every change. A background refresh
  // only applies its answer if no change started or finished while it ran.
  const inFlight = useRef(0);
  const changeSeq = useRef(0);

  useEffect(() => {
    itemsRef.current = items;
    try {
      localStorage.setItem(ITEMS_KEY, JSON.stringify(items));
    } catch {
      // storage blocked: cart just won't survive a reload
    }
  }, [items]);

  /**
   * Asks the server to make this cart hold `want` units of an offer. The server
   * reserves/releases the stock atomically. On "not enough stock" (409) the
   * server also says what we really hold, so the cart is corrected to match.
   */
  const sync = async (offerId: string, want: number): Promise<SyncResult> => {
    if (want > MAX_PER_LINE) return { ok: false, error: TOO_MANY };
    inFlight.current++;
    changeSeq.current++;
    try {
      const line = await api.setCartItem(cartId, offerId, want);
      setLiveStock((s) => ({ ...s, [offerId]: line.stock }));
      return { ok: true, quantity: line.quantity, stock: line.stock, expiresAt: line.expiresAt };
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        const body = e.body as { stock: number; quantity: number };
        setLiveStock((s) => ({ ...s, [offerId]: body.stock }));
        setItems((cur) => applyServerLine(cur, offerId, body.quantity, body.stock));
        return { ok: false, error: outOfStockMessage(body.stock, body.quantity) };
      }
      return { ok: false, error: FAILED };
    } finally {
      inFlight.current--;
      changeSeq.current++;
    }
  };

  /**
   * Reservations expire on the server after 30 min (DECISIONS #16). This asks
   * what the server still holds and drops expired lines, warning the shopper.
   */
  const refresh = useCallback(async () => {
    if (itemsRef.current.length === 0 || inFlight.current > 0) return;
    const seq = changeSeq.current;
    try {
      const { items: server } = await api.cart(cartId);
      if (seq !== changeSeq.current) return; // a change ran meanwhile; its answer wins
      const { items: next, lost } = reconcile(itemsRef.current, server);
      setItems(next);
      if (lost) setNotice({ kind: "error", title: "Krepšelis atnaujintas", message: EXPIRED });
    } catch {
      // offline or server down: try again on the next tick
    }
  }, [cartId]);

  // Check on load, every minute, and when the shopper comes back to the tab.
  useEffect(() => {
    void refresh();
    const timer = setInterval(refresh, 60_000);
    window.addEventListener("focus", refresh);
    return () => {
      clearInterval(timer);
      window.removeEventListener("focus", refresh);
    };
  }, [refresh]);

  const add = async (product: Product, offer: Offer) => {
    const res = await sync(offer.id, quantityOf(itemsRef.current, offer.id) + 1);
    if (!res.ok) {
      setNotice({ kind: "error", message: res.error });
      return;
    }
    setItems((cur) => addServerLine(cur, product, offer, res.quantity, res.stock, res.expiresAt));
    setNotice({ kind: "added" });
  };

  const setQty = async (offerId: string, qty: number): Promise<string | null> => {
    const res = await sync(offerId, qty);
    if (!res.ok) return res.error;
    setItems((cur) => applyServerLine(cur, offerId, res.quantity, res.stock, res.expiresAt));
    return null;
  };

  const closeNotice = useCallback(() => setNotice(null), []);

  const value: CartValue = {
    items,
    count: itemCount(items),
    total: totalCents(items),
    liveStock,
    add,
    setQty,
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
