import { useState } from "react";
import { Link } from "react-router-dom";
import { formatPrice } from "../../lib/format";
import type { Order, OrderStatus } from "../../types";

const STATUS: Record<OrderStatus, string> = {
  cart: "Krepšelis",
  paid: "Apmokėta",
  delivered: "Pristatyta",
};

/** One order with its key codes; used after checkout and in the order history. */
export function OrderCard({ order }: { order: Order }) {
  const date = new Date(order.paidAt ?? order.createdAt).toLocaleString("lt-LT");
  return (
    <section className="order">
      <header className="order-head">
        <span className="muted small">Užsakymas {order.id.slice(0, 8)} · {date}</span>
        <span className={`status status-${order.status}`}>{STATUS[order.status]}</span>
      </header>
      <ul className="order-items">
        {order.items.map((it) => (
          <li key={it.offerId} className="order-item">
            {it.imageUrl ? <img className="cart-img" src={it.imageUrl} alt="" /> : <div className="cart-img" />}
            <div className="cart-info">
              <Link to={`/product/${it.productId}`} className="cart-title">{it.productName}</Link>
              <span className="muted small">
                Pardavėjas: {it.sellerName} · {it.quantity} × {formatPrice(it.price, order.currency)}
              </span>
              {it.keys?.length ? (
                it.keys.map((k) => <KeyCode key={k.code} code={k.code} />)
              ) : (
                <span className="muted small">Raktai bus parodyti, kai užsakymas bus pristatytas.</span>
              )}
            </div>
          </li>
        ))}
      </ul>
      <div className="cart-total">
        <span>Iš viso</span>
        <span className="price">{formatPrice(order.total, order.currency)}</span>
      </div>
    </section>
  );
}

function KeyCode({ code }: { code: string }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
    } catch {
      // clipboard blocked: the code is still on screen to copy by hand
    }
  };
  return (
    <div className="key-row">
      <code className="key-code">{code}</code>
      <button type="button" className="link-btn" onClick={copy}>{copied ? "Nukopijuota" : "Kopijuoti"}</button>
    </div>
  );
}
