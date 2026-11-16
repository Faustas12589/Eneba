import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { formatDate, formatPrice } from "../../lib/format";
import { Status } from "../../components/Status";
import { useAuth } from "./AuthContext";
import { Avatar } from "./Avatar";

const PROVIDERS: Record<string, string> = {
  "google.com": "Google",
  password: "El. paštas ir slaptažodis",
  dev: "Kūrimo režimas",
};

/**
 * /account (signed in only). Name, photo, email, sign-in method and sign-up
 * date come from Firebase Auth in the browser; order count and total spent
 * come from the API (GET /api/me), because orders live on the server.
 */
export function AccountPage() {
  const { user, updateName } = useAuth();
  const stats = useAsync(() => api.me(), [user?.uid]);
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  if (!user) return null; // RequireAuth sends guests to /login first

  async function saveName(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await updateName(name.trim());
      setEditing(false);
    } catch {
      setError("Nepavyko išsaugoti vardo. Bandykite dar kartą.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="page container narrow">
      <h1>Mano paskyra</h1>
      <section className="panel account">
        <div className="account-head">
          <Avatar user={user} size={96} />
          <div className="account-head-text">
            {editing ? (
              <form className="account-name-form" onSubmit={saveName}>
                <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Vardas" maxLength={60} autoFocus />
                <button className="btn" disabled={busy}>Išsaugoti</button>
                <button type="button" className="btn secondary" onClick={() => setEditing(false)}>Atšaukti</button>
              </form>
            ) : (
              <h2>
                {user.name || "Vardas nenurodytas"}{" "}
                <button type="button" className="link-btn" onClick={() => { setName(user.name ?? ""); setEditing(true); }}>
                  Keisti
                </button>
              </h2>
            )}
            {error && <p className="error" role="alert">{error}</p>}
            <span className="muted">{user.email}</span>
          </div>
        </div>

        <dl className="account-facts">
          <dt>Prisijungimo būdas</dt>
          <dd>{PROVIDERS[user.provider] ?? user.provider}</dd>
          {user.createdAt && (
            <>
              <dt>Narys nuo</dt>
              <dd>{formatDate(user.createdAt)}</dd>
            </>
          )}
        </dl>
        <Status state={stats}>
          {(me) => (
            <dl className="account-facts">
              <dt>Užsakymai</dt>
              <dd>{me.orderCount}</dd>
              <dt>Iš viso išleista</dt>
              <dd className="price">{formatPrice(me.totalSpent)}</dd>
            </dl>
          )}
        </Status>
        <Link to="/orders" className="btn secondary">Mano užsakymai</Link>
      </section>
    </div>
  );
}
