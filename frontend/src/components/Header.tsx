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
          <Link to="/orders">Užsakymai</Link>
          {/* Login is off for now (DECISIONS #20). Only shown if someone signed in earlier. */}
          {user && <button className="btn secondary" onClick={signOut}>Atsijungti</button>}
        </nav>
      </div>
    </header>
  );
}
