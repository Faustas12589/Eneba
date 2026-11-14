import { Link } from "react-router-dom";
import { formatPrice } from "../../lib/format";
import { useCart } from "./CartContext";

export function CartPage() {
  const { items, total, setQty, remove } = useCart();

  if (items.length === 0) {
    return (
      <div className="page container">
        <h1>Cart</h1>
        <p className="muted">Your cart is empty.</p>
        <Link to="/store" className="btn">Browse the store</Link>
      </div>
    );
  }

  const currency = items[0].currency; // EUR only (SPEC assumptions)
  return (
    <div className="page container">
      <h1>Cart</h1>
      <div className="cart">
        <ul className="cart-lines">
          {items.map((i) => (
            <li key={i.offerId} className="cart-line">
              {i.imageUrl ? <img className="cart-img" src={i.imageUrl} alt="" /> : <div className="cart-img" />}
              <div className="cart-info">
                <Link to={`/product/${i.productId}`} className="card-title">{i.productName}</Link>
                <span className="muted">Seller: {i.sellerName}</span>
                <button className="link-btn" onClick={() => remove(i.offerId)}>Remove</button>
              </div>
              <input
                type="number"
                className="qty"
                min={1}
                max={i.stock}
                value={i.quantity}
                aria-label={`Quantity of ${i.productName}`}
                onChange={(e) => setQty(i.offerId, e.target.valueAsNumber)}
              />
              <span className="price">{formatPrice(i.price * i.quantity, i.currency)}</span>
            </li>
          ))}
        </ul>
        <aside className="cart-summary">
          <div className="cart-total">
            <span>Total</span>
            <span className="price">{formatPrice(total, currency)}</span>
          </div>
          <p className="muted">Final price is confirmed at checkout.</p>
          <Link to="/checkout" className="btn">Proceed to checkout</Link>
        </aside>
      </div>
    </div>
  );
}
