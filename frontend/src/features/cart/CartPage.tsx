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
      <div className="cart">
        <section className="panel">
          <h2>Mano krepšelis</h2>
          <ul className="cart-lines">
            {items.map((i) => <CartLine key={i.offerId} item={i} />)}
          </ul>
        </section>
        <aside className="cart-summary panel">
          <h2>Suvestinė</h2>
          <Link to="/checkout" className="btn block">Atsiskaityti dabar</Link>
          <div className="cart-total small">
            <span>{formatItemCount(count)}</span>
            <span>{formatPrice(total, currency)}</span>
          </div>
          <div className="cart-total cart-grand-total">
            <span>Iš viso:</span>
            <span className="price">{formatPrice(total, currency)}</span>
          </div>
        </aside>
      </div>
    </div>
  );
}
