import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { formatItemCount } from "../../lib/format";
import { useCart } from "./CartContext";

/** Modal shown after "Add to cart": success, or why the item couldn't be added. */
export function CartPopup() {
  const { notice, closeNotice, count } = useCart();
  const navigate = useNavigate();

  useEffect(() => {
    if (!notice) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") closeNotice();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [notice, closeNotice]);

  if (!notice) return null;

  const goToCart = () => {
    closeNotice();
    navigate("/cart");
  };

  return (
    <div className="modal-backdrop" onClick={closeNotice}>
      <div
        className="modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="cart-popup-title"
        onClick={(e) => e.stopPropagation()}
      >
        <button type="button" className="modal-close" aria-label="Uždaryti" onClick={closeNotice}>×</button>
        {notice.kind === "added" ? (
          <>
            <h2 id="cart-popup-title"><span className="ok-icon">✓</span> Pridėta į krepšelį</h2>
            <p className="muted">Krepšelyje yra {formatItemCount(count)}</p>
            <div className="modal-actions">
              <button type="button" className="btn secondary" onClick={closeNotice} autoFocus>Tęsti apsipirkimą</button>
              <button type="button" className="btn" onClick={goToCart}>Peržiūrėti krepšelį</button>
            </div>
          </>
        ) : (
          <>
            <h2 id="cart-popup-title">Nepavyko pridėti</h2>
            <p className="error">{notice.message}</p>
            <div className="modal-actions">
              <button type="button" className="btn" onClick={closeNotice} autoFocus>Gerai</button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
