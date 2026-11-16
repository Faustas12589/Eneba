import { Link, useLocation } from "react-router-dom";
import { useDropdown } from "../../lib/useDropdown";
import { useAuth } from "./AuthContext";
import { Avatar } from "./Avatar";

/** Header avatar + dropdown: who is signed in and their links, or a sign-in link for guests. */
export function AccountMenu() {
  const { user, signOut } = useAuth();
  const { open, setOpen, ref, close } = useDropdown();
  const { pathname } = useLocation();

  return (
    <div className="account-menu" ref={ref}>
      <button
        type="button"
        className="nav-btn avatar-btn"
        aria-expanded={open}
        aria-label="Paskyra"
        title={user ? user.name || user.email || "Paskyra" : "Prisijungti"}
        onClick={() => setOpen((o) => !o)}
      >
        <Avatar user={user} />
      </button>
      {open && (
        <div className="account-dropdown">
          {user ? (
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
          ) : (
            <>
              <Link to="/login" state={{ from: pathname }} className="btn block" onClick={close}>Prisijungti</Link>
              <Link to="/orders" onClick={close}>Mano užsakymai</Link>
            </>
          )}
        </div>
      )}
    </div>
  );
}
