import { Link } from "react-router-dom";
import type { Product } from "../types";
import { formatPrice, regionLabel } from "../lib/format";

/** Eneba-style card: cover, platform strip, then title, region, "Nuo" + old price, price. */
export function ProductCard({ p }: { p: Product }) {
  return (
    <Link to={`/product/${p.id}`} className="card">
      {p.imageUrl ? (
        <img className="card-img" src={p.imageUrl} alt={p.name} />
      ) : (
        <div className="card-img card-img-empty" aria-hidden="true">{p.name}</div>
      )}
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
  );
}
