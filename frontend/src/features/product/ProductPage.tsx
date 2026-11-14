import { useState } from "react";
import { useParams } from "react-router-dom";
import type { Offer, Product } from "../../types";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { formatPrice } from "../../lib/format";
import { Status } from "../../components/Status";
import { useCart } from "../cart/CartContext";

export function ProductPage() {
  const { id = "" } = useParams();
  const state = useAsync(() => api.product(id), [id]);
  const { items, add, liveStock } = useCart();
  const [pending, setPending] = useState<string | null>(null); // offer id being added
  const inCart = (offerId: string) => items.find((i) => i.offerId === offerId)?.quantity ?? 0;
  // Stock changes as carts reserve it; prefer the newest number the server gave us.
  const stockOf = (o: Offer) => liveStock[o.id] ?? o.stock;
  const onAdd = async (p: Product, o: Offer) => {
    setPending(o.id);
    await add(p, o); // reserves on the server; the popup shows success or the error
    setPending(null);
  };
  return (
    <div className="page container">
      <Status state={state}>
        {({ product: p, offers }) => (
          <div className="product">
            {p.imageUrl ? <img className="card-img" src={p.imageUrl} alt={p.name} /> : <div className="card-img" />}
            <div>
              <h1>{p.name}</h1>
              <p className="muted">{p.platform} · {p.region} · {p.genre}</p>
              <p>{p.description}</p>
              <h2>Pardavėjų pasiūlymai</h2>
              <table className="offers">
                <thead>
                  <tr><th>Pardavėjas</th><th>Įvertinimas</th><th>Likutis</th><th>Kaina</th><th /></tr>
                </thead>
                <tbody>
                  {offers.map((o, i) => (
                    <tr key={o.id}>
                      <td>{o.sellerName} {i === 0 && <span className="badge">Geriausia kaina</span>}</td>
                      <td>★ {o.sellerRating.toFixed(1)}</td>
                      <td>{stockOf(o)} vnt.</td>
                      <td className="price">{formatPrice(o.price, o.currency)}</td>
                      <td>
                        <button
                          className="btn"
                          disabled={stockOf(o) < 1 || pending === o.id}
                          onClick={() => onAdd(p, o)}
                        >
                          {stockOf(o) < 1 ? "Išparduota" : "Į krepšelį"}
                        </button>
                        {inCart(o.id) > 0 && <span className="muted"> {inCart(o.id)} krepšelyje</span>}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </Status>
    </div>
  );
}
