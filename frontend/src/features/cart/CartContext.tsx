import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import { api, ApiError } from "../../lib/api";
import type { Order, OrderItem, OutOfStock } from "../../types";
import { emptyCart, itemCount, lostLines, MAX_PER_LINE, outOfStockMessage, quantityOf } from "./cart";

/** What the add-to-cart popup shows: success, or why the add failed. */
export type CartNotice = { kind: "added" } | { kind: "error"; title?: string; message: string };

interface CartValue {
  /** The cart order as the server stores it (DECISIONS #21). */
  items: OrderItem[];
  count: number; // total quantity, for the header badge and popup
  total: number; // cents, from the server
  currency: string;
  /** Latest stock the server reported per offer id (product page table, cart lines). */
  liveStock: Record<string, number>;
  /** Reserves one more unit on the server, then opens the popup (success or error). */
  add: (offerId: string) => Promise<void>;
  /**
   * Card hover buttons: adds one unit of the product's cheapest offer in stock.
   * `quiet` skips the success popup ("Pirkti dabar" goes straight to checkout).
   * Errors still open the popup. Returns whether a unit was added.
   */
  addProduct: (productId: string, quiet?: boolean) => Promise<boolean>;
  /** Sets a line's quantity on the server (0 = remove). Returns an error message or null. */
  setQty: (offerId: string, qty: number) => Promise<string | null>;
  /** Simulated payment; starts a new cart and returns the delivered order. Throws on failure. */
  checkout: () => Promise<Order>;
  notice: CartNotice | null;
  closeNotice: () => void;
}

const Ctx = createContext<CartValue | null>(null);
const ID_KEY = "cartId";
const FAILED = "Nepavyko atnaujinti krepšelio. Bandykite dar kartą.";
const EXPIRED = "Rezervacija baigėsi: kai kurios prekės grąžintos į sandėlį ir pašalintos iš krepšelio.";
const CLOSED = "Šis krepšelis jau apmokėtas, todėl pradėtas naujas. Bandykite dar kartą.";
const TOO_MANY = `Vienam pasiūlymui galima daugiausia ${MAX_PER_LINE} vnt.`;

/** Random id that names this browser's cart on the server (like a guest session). */
function loadCartId(): string {
  try {
    return localStorage.getItem(ID_KEY) ?? newCartId();
  } catch {
    return crypto.randomUUID(); // storage blocked: cart lives for this page load only
  }
}

/** A fresh cart id. The old one now names a paid order, which can't change. */
function newCartId(): string {
  const id = crypto.randomUUID();
  try {
    localStorage.setItem(ID_KEY, id);
  } catch {
    // storage blocked: the new id lasts for this page load only
  }
  return id;
}

export function CartProvider({ children }: { children: React.ReactNode }) {
  // The id lives in a ref: it changes after checkout and async code must see
  // the new one. useState just reads localStorage once instead of every render.
  const [initialCartId] = useState(loadCartId);
  const cartId = useRef(initialCartId);
  // The server's cart document. Only the id is kept in the browser; lines,
  // names, prices and the total always come from the server.
  const [cart, setCart] = useState<Order>(() => emptyCart(initialCartId));
  const loaded = useRef(false); // false until the first GET answered
  const [liveStock, setLiveStock] = useState<Record<string, number>>({});
  const [notice, setNotice] = useState<CartNotice | null>(null);
  // Latest cart for async functions: after `await`, a closed-over value may be old.
  const cartRef = useRef(cart);
  cartRef.current = cart;
  // Changes in flight + a counter bumped by every change. A background refresh
  // only applies its answer if no change started or finished while it ran.
  const inFlight = useRef(0);
  const changeSeq = useRef(0);

  /** Forget the paid/closed cart and use a new server cart id. */
  function startNewCart() {
    cartId.current = newCartId();
    setCart(emptyCart(cartId.current));
  }

  /**
   * Reads the cart from the server. Reservations expire after 30 min
   * (DECISIONS #16): if lines are gone, the shopper gets a warning.
   */
  const refresh = useCallback(async () => {
    if (inFlight.current > 0) return;
    if (loaded.current && cartRef.current.items.length === 0) return; // nothing to expire
    const seq = changeSeq.current;
    try {
      const next = await api.cart(cartId.current);
      if (seq !== changeSeq.current) return; // a change ran meanwhile; its answer wins
      if (loaded.current && lostLines(cartRef.current.items, next.items)) {
        setNotice({ kind: "error", title: "Krepšelis atnaujintas", message: EXPIRED });
      }
      loaded.current = true;
      setCart(next);
    } catch (e) {
      if (e instanceof ApiError && e.status === 410) {
        loaded.current = true;
        startNewCart(); // paid in another tab: it's an order now
      }
      // offline or server down: try again on the next tick
    }
  }, []);

  // Load on start, then check every minute and when the shopper comes back to the tab.
  useEffect(() => {
    void refresh();
    const timer = setInterval(refresh, 60_000);
    window.addEventListener("focus", refresh);
    return () => {
      clearInterval(timer);
      window.removeEventListener("focus", refresh);
    };
  }, [refresh]);

  /**
   * Runs one cart change on the server and shows its answer. Counted as a
   * change, so a background refresh started earlier can't overwrite it.
   */
  const change = async <T,>(run: () => Promise<T>): Promise<T> => {
    inFlight.current++;
    changeSeq.current++;
    try {
      return await run();
    } finally {
      inFlight.current--;
      changeSeq.current++;
    }
  };

  /**
   * Asks the server to make this cart hold `want` units of an offer. The server
   * reserves/releases stock and keys atomically and returns the stored cart.
   * Returns an error message, or null on success.
   */
  const sync = (offerId: string, want: number): Promise<string | null> =>
    change(async () => {
      if (want > MAX_PER_LINE) return TOO_MANY;
      try {
        const res = await api.setCartItem(cartId.current, offerId, want);
        setCart(res.cart);
        setLiveStock((s) => ({ ...s, [offerId]: res.stock }));
        return null;
      } catch (e) {
        if (e instanceof ApiError && e.status === 410) {
          startNewCart();
          return CLOSED;
        }
        if (e instanceof ApiError && e.status === 409) {
          const body = e.body as OutOfStock;
          setLiveStock((s) => ({ ...s, [offerId]: body.stock }));
          // Show what the server really holds now.
          await api.cart(cartId.current).then(setCart, () => {});
          return outOfStockMessage(body.stock, body.quantity);
        }
        return FAILED;
      }
    });

  const addOne = async (offerId: string, quiet: boolean) => {
    const error = await sync(offerId, quantityOf(cartRef.current.items, offerId) + 1);
    if (error) setNotice({ kind: "error", message: error });
    else if (!quiet) setNotice({ kind: "added" });
    return !error;
  };
  const add = async (offerId: string) => void (await addOne(offerId, false));

  const addProduct = async (productId: string, quiet = false) => {
    let offerId: string | undefined;
    try {
      // The API lists offers cheapest first; the server re-checks stock when reserving.
      offerId = (await api.product(productId)).offers.find((o) => o.stock > 0)?.id;
    } catch {
      setNotice({ kind: "error", message: FAILED });
      return false;
    }
    if (!offerId) {
      setNotice({ kind: "error", title: "Išparduota", message: "Šios prekės šiuo metu neturi nė vienas pardavėjas." });
      return false;
    }
    return addOne(offerId, quiet);
  };

  const setQty = (offerId: string, qty: number) => sync(offerId, qty);

  /**
   * Pays for the cart (simulated) and returns the delivered order with its keys.
   * Counts as a change, so a background refresh can't see the now-paid cart and
   * show a wrong warning. Throws ApiError on failure.
   */
  const checkout = () =>
    change(async () => {
      const order = await api.checkout(cartId.current);
      startNewCart();
      return order;
    });

  const closeNotice = useCallback(() => setNotice(null), []);

  const value: CartValue = {
    items: cart.items,
    count: itemCount(cart.items),
    total: cart.total,
    currency: cart.currency,
    liveStock,
    add,
    addProduct,
    setQty,
    checkout,
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
