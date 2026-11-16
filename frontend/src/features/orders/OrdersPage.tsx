import { Link, useSearchParams } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import { Pager } from "../../components/Pager";
import { OrderCard } from "./OrderCard";
import { useAuth } from "../auth/AuthContext";

/**
 * The user's past orders (or this browser's, for a guest), newest first, paged,
 * with their key codes. Page number lives in the URL. Reloads on sign-in/out.
 */
export function OrdersPage() {
  const [sp, setSp] = useSearchParams();
  const page = Number(sp.get("page")) || 1;
  const uid = useAuth().user?.uid;
  const state = useAsync((signal) => api.orders(page, signal), [page, uid]);

  function goTo(n: number) {
    setSp(n > 1 ? { page: String(n) } : {});
    window.scrollTo(0, 0);
  }

  return (
    <div className="page container narrow">
      <h1>Mano užsakymai</h1>
      <Status state={state}>
        {(pg) =>
          pg.total === 0 ? (
            <>
              <p className="muted">Užsakymų dar nėra.</p>
              <Link to="/store" className="btn">Į parduotuvę</Link>
            </>
          ) : (
            <>
              <div className="orders">
                {pg.items.map((o) => <OrderCard key={o.id} order={o} />)}
              </div>
              {pg.items.length === 0 && <p className="muted">Šiame puslapyje užsakymų nėra.</p>}
              <Pager page={pg.page} pageSize={pg.pageSize} total={pg.total} onPage={goTo} />
            </>
          )
        }
      </Status>
    </div>
  );
}
