import { Link, useLocation } from "react-router-dom";
import { useDropdown } from "../../lib/useDropdown";
import { useAuth } from "./AuthContext";
import { Avatar } from "./Avatar";
import { UserIcon } from "../../components/Icons";

/** Header account area: sign-in/register links for guests; avatar + dropdown when signed in. */
export function AccountMenu() {
  const { user, signOut } = useAuth();
  const { open, setOpen, ref, close } = useDropdown();
  const { pathname } = useLocation();

  // Signed out: Eneba's "👤 Prisijungti | Registruotis" text links, no dropdown.
  if (!user) {
    return (
      <div className="auth-links">
        <UserIcon size={26} />
        <Link to="/login" state={{ from: pathname }}>Prisijungti</Link>
        <span className="auth-sep" aria-hidden="true">|</span>
        <Link to="/login?mode=register" state={{ from: pathname }}>Registruotis</Link>
      </div>
    );
  }

  return (
    <div className="account-menu" ref={ref}>
      <button
        type="button"
        className="nav-btn avatar-btn"
        aria-expanded={open}
        aria-label="Paskyra"
        title={user.name || user.email || "Paskyra"}
        onClick={() => setOpen((o) => !o)}
      >
        <Avatar user={user} />
      </button>
      {open && (
        <div className="account-dropdown">
          <>
              <div className="account-who">
                <Avatar user={user} size={48} />
                <div className="account-who-text">
                  {user.name && <strong>{user.name}</strong>}
                  <span className="muted small">{user.email}</span>
                </div>
              </div>
              <Link to="/account" onClick={close}>Mano paskyra</Link>
              <Link to="/orders" onClick={close}>Mano užsakymai</Link>
              <button type="button" className="account-signout" onClick={() => { close(); signOut(); }}>
                Atsijungti
              </button>
          </>
        </div>
      )}
    </div>
  );
}
