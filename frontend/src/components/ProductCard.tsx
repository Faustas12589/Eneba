import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import type { Product } from "../types";
import { formatPrice, regionLabel } from "../lib/format";
import { useCart } from "../features/cart/CartContext";
import { Cover } from "./Cover";
import { FavoriteButton } from "../features/favorites/FavoriteButton";

/**
 * Eneba-style card: cover, platform strip, then title, region, "Nuo" + old price, price.
 * On hover (or keyboard focus) the info panel grows over the cover and shows
 * "Pridėti į krepšelį" and "Pirkti dabar"; both use the cheapest offer in stock.
 * The buttons sit outside the link, because a button inside a link is invalid HTML.
 */
export function ProductCard({ p }: { p: Product }) {
  const { addProduct } = useCart();
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);

  async function run(buyNow: boolean) {
    setBusy(true);
    try {
      const added = await addProduct(p.id, buyNow);
      if (added && buyNow) navigate("/checkout");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="card">
      <Link to={`/product/${p.id}`} className="card-link">
        {p.preOrder && <span className="card-tag" title="Išankstinis užsakymas">Iš anksto · {p.releaseDate}</span>}
        <Cover className="card-img" productId={p.id} name={p.name} imageUrl={p.imageUrl} category={p.category} />
        <div className="card-platform">{p.platform}</div>
        <div className="card-body">
          <div className="card-title">{p.name}</div>
          <div className="card-region">{regionLabel(p.region)}</div>
          <div className="card-from">
            Nuo{" "}
            {p.discountPercent > 0 && (
              <>
                <s>{formatPrice(p.listPrice, p.currency)}</s> <span className="discount">-{p.discountPercent}%</span>
              </>
            )}
          </div>
          <div className="price">{formatPrice(p.priceFrom, p.currency)}</div>
        </div>
      </Link>
      {/* outside the link like the buttons below: a button inside a link is invalid HTML */}
      <FavoriteButton productId={p.id} />
      <div className="card-actions">
        <button type="button" className="btn block" disabled={busy} onClick={() => run(false)}>Pridėti į krepšelį</button>
        <button type="button" className="btn outline block" disabled={busy} onClick={() => run(true)}>Pirkti dabar</button>
      </div>
    </div>
  );
}
