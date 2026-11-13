import { Link } from "react-router-dom";
import type { Product } from "../types";
import { formatPrice } from "../lib/format";

export function ProductCard({ p }: { p: Product }) {
  return (
    <Link to={`/product/${p.id}`} className="card">
      {p.imageUrl ? <img className="card-img" src={p.imageUrl} alt={p.name} /> : <div className="card-img" />}
      <div className="card-body">
        <div className="card-title">{p.name}</div>
        <div className="muted">{p.platform} · {p.region}</div>
        <div>
          {p.discountPercent > 0 && <span className="badge">-{p.discountPercent}%</span>}{" "}
          <span className="price">{formatPrice(p.priceFrom, p.currency)}</span>
        </div>
      </div>
    </Link>
  );
}
