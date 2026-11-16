import { useState } from "react";
import { Link } from "react-router-dom";
import { ApiError } from "../../lib/api";
import { formatItemCount, formatPrice } from "../../lib/format";
import { useCart } from "../cart/CartContext";
import { useAuth } from "../auth/AuthContext";
import { SignInModal } from "../auth/SignInModal";
import { OrderCard } from "../orders/OrderCard";
import type { Order } from "../../types";

/** Why checkout failed, in words the shopper can act on. */
function checkoutError(e: unknown): string {
  if (e instanceof ApiError && e.status === 400) return "Krepšelis tuščias arba rezervacija baigėsi. Grįžkite į krepšelį.";
  if (e instanceof ApiError && e.status === 401) return "Prisijungimas baigėsi. Prisijunkite ir bandykite dar kartą.";
  if (e instanceof ApiError && e.status === 409) return "Krepšelis pasikeitė apmokant. Bandykite dar kartą.";
  return "Nepavyko apmokėti. Bandykite dar kartą: dvigubai nenuskaičiuosime.";
}

/**
 * Simulated payment (DECISIONS #8). The server turns the cart into an order,
 * cart -> paid -> delivered, and returns the keys that were reserved for it.
 * Paying is the one step that needs an account: a guest gets the sign-in
 * popup, and payment continues once they are in (DECISIONS #28).
 */
export function CheckoutPage() {
  const { items, count, total, currency, checkout } = useCart();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [order, setOrder] = useState<Order | null>(null);
  const { user } = useAuth();
  const [askSignIn, setAskSignIn] = useState(false);

  const pay = async () => {
    setBusy(true);
    setError(null);
    try {
      setOrder(await checkout());
    } catch (e) {
      setError(checkoutError(e));
    } finally {
      setBusy(false);
    }
  };

  if (order) {
    return (
      <div className="page container narrow">
        <h1><span className="ok-icon">✓</span> Ačiū! Užsakymas pristatytas</h1>
        <p className="muted">Jūsų raktai žemiau. Juos visada rasite skiltyje „Mano užsakymai“.</p>
        <OrderCard order={order} />
        <div className="modal-actions">
          <Link to="/orders" className="btn">Mano užsakymai</Link>
          <Link to="/store" className="btn secondary">Pirkti toliau</Link>
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="page container">
        <h1>Atsiskaitymas</h1>
        <p className="muted">Krepšelis tuščias.</p>
        <Link to="/store" className="btn">Į parduotuvę</Link>
      </div>
    );
  }

  return (
    <div className="page container narrow">
      <h1>Atsiskaitymas</h1>
      <section className="order">
        <ul className="order-items">
          {items.map((i) => (
            <li key={i.offerId} className="checkout-line">
              <span>{i.productName} <span className="muted small">· {i.sellerName}</span></span>
              <span>{i.quantity} × {formatPrice(i.price, currency)}</span>
            </li>
          ))}
        </ul>
        <div className="cart-total">
          <span>Suma ({formatItemCount(count)})</span>
          <span className="price">{formatPrice(total, currency)}</span>
        </div>
        <p className="muted small">
          Mokėjimas imituojamas: pinigai nenuskaičiuojami. Apmokėjus rezervuoti raktai iškart tampa jūsų.
        </p>
        {error && <p className="error" role="alert">{error}</p>}
        <button type="button" className="btn block" disabled={busy} onClick={user ? pay : () => setAskSignIn(true)}>
          {busy ? "Apmokama…" : `Apmokėti ${formatPrice(total, currency)}`}
        </button>
        <Link to="/cart" className="link-btn">Grįžti į krepšelį</Link>
      </section>
      {askSignIn && (
        <SignInModal
          onClose={() => setAskSignIn(false)}
          onSignedIn={() => {
            setAskSignIn(false);
            pay();
          }}
        />
      )}
    </div>
  );
}
