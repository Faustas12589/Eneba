import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import { HeartIcon } from "../../components/Icons";
import { formatItemCount } from "../../lib/format";
import { useAuth } from "../auth/AuthContext";
import { useFavorites } from "./FavoritesContext";

/**
 * "Norų sąrašas": the user's favorite products, newest first. Guests have one too
 * (kept for this browser, moved to the account on sign-in), like on Eneba.
 * Un-hearting a card hides it at once: the list is filtered by the live ids.
 */
export function FavoritesPage() {
  const { user, ready } = useAuth();
  const { ids } = useFavorites();
  const state = useAsync(() => api.favorites(), [user?.uid, ready]);

  return (
    <div className="page container">
      <h1>Norų sąrašas</h1>
      {!user && ready && (
        <p className="muted">
          Sąrašas saugomas šioje naršyklėje. <Link to="/login">Prisijunk</Link>, ir jį matysi bet kuriame įrenginyje.
        </p>
      )}
      <Status state={state}>
        {({ items }) => {
          const shown = items.filter((p) => ids.includes(p.id));
          return shown.length === 0 ? (
            <div className="empty-favorites">
              <HeartIcon size={48} />
              <p>Norų sąrašas tuščias. Spausk širdelę ant prekės, kad ją išsaugotum.</p>
              <Link to="/store" className="btn">Į parduotuvę</Link>
            </div>
          ) : (
            <>
              <p className="muted">{formatItemCount(shown.length)}</p>
              <div className="grid">{shown.map((p) => <ProductCard key={p.id} p={p} />)}</div>
            </>
          );
        }}
      </Status>
    </div>
  );
}
