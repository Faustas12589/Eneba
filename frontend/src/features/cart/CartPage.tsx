import { Link } from "react-router-dom";
import { formatItemCount, formatPrice } from "../../lib/format";
import { useCart } from "./CartContext";
import { CartLine } from "./CartLine";

export function CartPage() {
  const { items, count, total, currency } = useCart();

  if (items.length === 0) {
    return (
      <div className="page container">
        <h1>Krepšelis</h1>
        <p className="muted">Krepšelis tuščias.</p>
        <Link to="/store" className="btn">Į parduotuvę</Link>
      </div>
    );
  }

  return (
    <div className="page container">
      <h1>Krepšelis</h1>
      <div className="cart">
        <ul className="cart-lines">
          {items.map((i) => <CartLine key={i.offerId} item={i} />)}
        </ul>
        <aside className="cart-summary">
          <div className="cart-total">
            <span>Suma ({formatItemCount(count)})</span>
            <span className="price">{formatPrice(total, currency)}</span>
          </div>
          <p className="muted small">Galutinė kaina patvirtinama atsiskaitant.</p>
          <Link to="/checkout" className="btn block">Atsiskaityti dabar</Link>
        </aside>
      </div>
    </div>
  );
}
