import { useState } from "react";
import { useCart } from "./CartContext";
import type { CartItem } from "./cart";

/** − qty + control. "+" asks the server for live stock and shows an error if none is left. */
export function QtyStepper({ item }: { item: CartItem }) {
  const { setQty, increase } = useCart();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onMinus = () => {
    setError(null);
    setQty(item.offerId, item.quantity - 1);
  };
  const onPlus = async () => {
    setBusy(true);
    setError(await increase(item.offerId));
    setBusy(false);
  };

  return (
    <div className="qty-wrap">
      <div className="stepper">
        <button type="button" aria-label="Mažinti kiekį" disabled={item.quantity <= 1} onClick={onMinus}>−</button>
        <span aria-live="polite">{item.quantity}</span>
        <button type="button" aria-label="Didinti kiekį" disabled={busy} onClick={onPlus}>+</button>
      </div>
      {error && <p className="error small" role="alert">{error}</p>}
    </div>
  );
}
