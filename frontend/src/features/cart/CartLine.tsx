import { Link } from "react-router-dom";
import { formatPrice } from "../../lib/format";
import { useCart } from "./CartContext";
import { QtyStepper } from "./QtyStepper";
import type { CartItem } from "./cart";

/** One cart row; used by both the cart page and the header dropdown (CSS sizes it). */
export function CartLine({ item }: { item: CartItem }) {
  const { remove } = useCart();
  const overStock = item.quantity > item.stock; // stock dropped after it was added
  return (
    <li className="cart-line">
      {item.imageUrl ? <img className="cart-img" src={item.imageUrl} alt="" /> : <div className="cart-img" />}
      <div className="cart-info">
        <Link to={`/product/${item.productId}`} className="card-title">{item.productName}</Link>
        <span className="muted small">Pardavėjas: {item.sellerName}</span>
        <span className={overStock ? "error small" : "muted small"}>
          Likutis: {item.stock} vnt.{overStock && " Sumažinkite kiekį."}
        </span>
        <button type="button" className="link-btn" onClick={() => remove(item.offerId)}>Pašalinti</button>
      </div>
      <QtyStepper item={item} />
      <span className="price">{formatPrice(item.price * item.quantity, item.currency)}</span>
    </li>
  );
}
