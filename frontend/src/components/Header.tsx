import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { useAuth } from "../features/auth/AuthContext";
import { CartMenu } from "../features/cart/CartMenu";
import { CATEGORIES } from "../lib/categories";
import { MenuIcon, UserIcon } from "./Icons";

export function Header() {
  const { user, signOut } = useAuth();
  return (
    <header className="header">
      <div className="container header-main">
        <Link to="/" className="logo">eneba<span>style</span></Link>
        <SearchBox />
        <nav className="nav">
          <CartMenu />
          <Link to="/orders" className="avatar" title="Mano užsakymai" aria-label="Mano užsakymai">
            <UserIcon size={22} />
          </Link>
          {/* Login is off for now (DECISIONS #20). Only shown if someone signed in earlier. */}
          {user && <button className="btn secondary" onClick={signOut}>Atsijungti</button>}
        </nav>
      </div>
      {/* Category row like Eneba's second header line; each link is a store filter. */}
      <div className="category-bar">
        <nav className="container category-nav" aria-label="Kategorijos">
          <Link to="/store" className="category-all"><MenuIcon size={22} /> Kategorijos</Link>
          {CATEGORIES.map((c) => (
            <Link key={c.id} to={`/store?category=${c.id}`}>{c.label}</Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
