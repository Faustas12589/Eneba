import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { useAuth } from "../features/auth/AuthContext";
import { CartMenu } from "../features/cart/CartMenu";

export function Header() {
  const { user, signOut } = useAuth();
  return (
    <header className="header">
      <div className="container">
        <Link to="/" className="logo">ENEBA-STYLE</Link>
        <SearchBox />
        <nav className="nav">
          <Link to="/store">Parduotuvė</Link>
          <CartMenu />
          {user ? (
            <>
              <Link to="/orders">Užsakymai</Link>
              <button className="btn secondary" onClick={signOut}>Atsijungti</button>
            </>
          ) : (
            <Link to="/login" className="btn">Prisijungti</Link>
          )}
        </nav>
      </div>
    </header>
  );
}
