import { useState } from "react";
import { Link } from "react-router-dom";
import { formatPrice, formatTime } from "../../lib/format";
import { useCart } from "./CartContext";
import type { OrderItem } from "../../types";

/**
 * One cart row; used by both the cart page and the header dropdown (CSS sizes it).
 * Every change goes to the server first (it reserves/releases stock), so the
 * buttons are disabled while a request is in flight.
 */
export function CartLine({ item }: { item: OrderItem }) {
  const { setQty, liveStock, currency } = useCart();
  const stock = liveStock[item.offerId]; // known once this browser changed the line
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
        <span className="muted small">Parduoda <strong>{item.sellerName}</strong></span>
        <Link to={`/product/${item.productId}`} className="cart-title">{item.productName}</Link>
        {stock !== undefined && <span className="muted small">Dar liko: {stock} vnt.</span>}
        {item.expiresAt && <span className="muted small">Rezervuota iki {formatTime(item.expiresAt)}</span>}
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
      <span className="price">{formatPrice(item.price * item.quantity, currency)}</span>
    </li>
  );
}
