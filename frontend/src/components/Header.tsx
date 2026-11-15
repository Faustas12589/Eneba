import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { useAuth } from "../features/auth/AuthContext";
import { CartMenu } from "../features/cart/CartMenu";
import { CATEGORIES } from "../lib/categories";

export function Header() {
  const { user, signOut } = useAuth();
  return (
    <header className="header">
      <div className="container header-main">
        <Link to="/" className="logo">eneba<span>style</span></Link>
        <SearchBox />
        <nav className="nav">
          <Link to="/store">Parduotuvė</Link>
          <CartMenu />
          <Link to="/orders">Užsakymai</Link>
          {/* Login is off for now (DECISIONS #20). Only shown if someone signed in earlier. */}
          {user && <button className="btn secondary" onClick={signOut}>Atsijungti</button>}
        </nav>
      </div>
      {/* Category row like Eneba's second header line; each link is a store filter. */}
      <div className="category-bar">
        <nav className="container category-nav" aria-label="Kategorijos">
          {CATEGORIES.map((c) => (
            <Link key={c.id} to={`/store?category=${c.id}`}>{c.label}</Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
