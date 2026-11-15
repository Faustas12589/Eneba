import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import { OrderCard } from "./OrderCard";

/** The signed-in user's past orders (newest first) with their key codes. */
export function OrdersPage() {
  const state = useAsync((signal) => api.orders(signal), []);
  return (
    <div className="page container narrow">
      <h1>Mano užsakymai</h1>
      <Status state={state}>
        {({ items }) =>
          items.length === 0 ? (
            <>
              <p className="muted">Užsakymų dar nėra.</p>
              <Link to="/store" className="btn">Į parduotuvę</Link>
            </>
          ) : (
            <div className="orders">
              {items.map((o) => <OrderCard key={o.id} order={o} />)}
            </div>
          )
        }
      </Status>
    </div>
  );
}
