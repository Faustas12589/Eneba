import { useParams } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { formatPrice } from "../../lib/format";
import { Status } from "../../components/Status";

export function ProductPage() {
  const { id = "" } = useParams();
  const state = useAsync(() => api.product(id), [id]);
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
              <h2>Offers from sellers</h2>
              <table className="offers">
                <thead>
                  <tr><th>Seller</th><th>Rating</th><th>Stock</th><th>Price</th><th /></tr>
                </thead>
                <tbody>
                  {offers.map((o, i) => (
                    <tr key={o.id}>
                      <td>{o.sellerName} {i === 0 && <span className="badge">Best price</span>}</td>
                      <td>★ {o.sellerRating.toFixed(1)}</td>
                      <td>{o.stock}</td>
                      <td className="price">{formatPrice(o.price, o.currency)}</td>
                      {/* TODO (cart slice): add this offer to the cart */}
                      <td><button className="btn" disabled>Add to cart</button></td>
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
