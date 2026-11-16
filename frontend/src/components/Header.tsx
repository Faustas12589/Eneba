import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { AccountMenu } from "../features/auth/AccountMenu";
import { CartMenu } from "../features/cart/CartMenu";
import { CATEGORIES } from "../lib/categories";
import { MenuIcon } from "./Icons";

export function Header() {
  return (
    <header className="header">
      <div className="container header-main">
        <Link to="/" className="logo">eneba<span>style</span></Link>
        <SearchBox />
        <nav className="nav">
          <CartMenu />
          <AccountMenu />
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
