import { Link } from "react-router-dom";
import { useDropdown } from "../../lib/useDropdown";
import { formatPrice } from "../../lib/format";
import { useCart } from "./CartContext";
import { CartLine } from "./CartLine";
import { CartIcon } from "../../components/Icons";

/** Header cart button + dropdown with the cart's lines, sum and two actions. */
export function CartMenu() {
  const { items, count, total, currency } = useCart();
  const { open, setOpen, ref, close } = useDropdown();

  return (
    <div className="cart-menu" ref={ref}>
      <button type="button" className="nav-btn cart-btn" aria-expanded={open} aria-label={`Krepšelis (${count})`} title="Krepšelis" onClick={() => setOpen((o) => !o)}>
        <CartIcon size={28} />
        {count > 0 && <span className="cart-count">{count}</span>}
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
                  <span className="price">{formatPrice(total, currency)}</span>
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
