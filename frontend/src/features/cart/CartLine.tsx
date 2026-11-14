import { useState } from "react";
import { Link } from "react-router-dom";
import { formatPrice } from "../../lib/format";
import { useCart } from "./CartContext";
import type { CartItem } from "./cart";

/**
 * One cart row; used by both the cart page and the header dropdown (CSS sizes it).
 * Every change goes to the server first (it reserves/releases stock), so the
 * buttons are disabled while a request is in flight.
 */
export function CartLine({ item }: { item: CartItem }) {
  const { setQty } = useCart();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const change = async (qty: number) => {
    setBusy(true);
    setError(await setQty(item.offerId, qty));
    setBusy(false);
  };

  return (
    <li className="cart-line">
      {item.imageUrl ? <img className="cart-img" src={item.imageUrl} alt="" /> : <div className="cart-img" />}
      <div className="cart-info">
        <Link to={`/product/${item.productId}`} className="card-title">{item.productName}</Link>
        <span className="muted small">Pardavėjas: {item.sellerName}</span>
        <span className="muted small">Dar liko: {item.stock} vnt.</span>
        <button type="button" className="link-btn" disabled={busy} onClick={() => change(0)}>Pašalinti</button>
      </div>
      <div className="qty-wrap">
        <div className="stepper">
          <button type="button" aria-label="Mažinti kiekį" disabled={busy || item.quantity <= 1} onClick={() => change(item.quantity - 1)}>−</button>
          <span aria-live="polite">{item.quantity}</span>
          <button type="button" aria-label="Didinti kiekį" disabled={busy} onClick={() => change(item.quantity + 1)}>+</button>
        </div>
        {error && <p className="error small" role="alert">{error}</p>}
      </div>
      <span className="price">{formatPrice(item.price * item.quantity, item.currency)}</span>
    </li>
  );
}
