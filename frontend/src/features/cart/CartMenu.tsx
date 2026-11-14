import { useEffect, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { formatPrice } from "../../lib/format";
import { useCart } from "./CartContext";
import { CartLine } from "./CartLine";

/** Header cart button + dropdown with the cart's lines, sum and two actions. */
export function CartMenu() {
  const { items, count, total } = useCart();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const { pathname } = useLocation();
  const close = () => setOpen(false);

  // Close when navigating (e.g. clicking a product name inside the dropdown).
  useEffect(() => setOpen(false), [pathname]);

  // Close on a click outside or Escape.
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className="cart-menu" ref={ref}>
      <button type="button" className="nav-btn" aria-expanded={open} onClick={() => setOpen((o) => !o)}>
        Krepšelis{count > 0 && <span className="cart-count">{count}</span>}
      </button>
      {open && (
        <div className="cart-dropdown">
          {items.length === 0 ? (
            <p className="muted">Krepšelis tuščias.</p>
          ) : (
            <>
              <ul className="cart-lines">
                {items.map((i) => <CartLine key={i.offerId} item={i} />)}
              </ul>
              <div className="cart-dropdown-footer">
                <div className="cart-total">
                  <span>Suma</span>
                  <span className="price">{formatPrice(total, items[0].currency)}</span>
                </div>
                <Link to="/cart" className="btn secondary block" onClick={close}>Peržiūrėti krepšelį</Link>
                <Link to="/checkout" className="btn block" onClick={close}>Atsiskaityti dabar</Link>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
