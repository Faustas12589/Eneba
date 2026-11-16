import { useState } from "react";
import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { AccountMenu } from "../features/auth/AccountMenu";
import { CartMenu } from "../features/cart/CartMenu";
import { CATEGORIES } from "../lib/categories";
import { CategoryDrawer } from "./CategoryDrawer";
import { HeartIcon, MenuIcon } from "./Icons";
import { useFavorites } from "../features/favorites/FavoritesContext";

export function Header() {
  const [drawer, setDrawer] = useState(false);
  const favorites = useFavorites().ids.length;
  return (
    <header className="header">
      <div className="container header-main">
        <Link to="/" className="logo">eneba<span>style</span></Link>
        <SearchBox />
        <nav className="nav">
          <Link to="/favorites" className="nav-btn icon-badge-btn" aria-label={`Norų sąrašas (${favorites})`} title="Norų sąrašas">
            <HeartIcon size={28} />
            {favorites > 0 && <span className="cart-count">{favorites}</span>}
          </Link>
          <CartMenu />
          <AccountMenu />
        </nav>
      </div>
      {/* Category row like Eneba's second header line; "Kategorijos" opens the side menu. */}
      <div className="category-bar">
        <nav className="container category-nav" aria-label="Kategorijos">
          <button type="button" className="category-all" aria-expanded={drawer} onClick={() => setDrawer(true)}>
            <MenuIcon size={22} /> Kategorijos
          </button>
          {CATEGORIES.map((c) => (
            <Link key={c.id} to={`/store?category=${c.id}`}>{c.label}</Link>
          ))}
        </nav>
      </div>
      {drawer && <CategoryDrawer onClose={() => setDrawer(false)} />}
    </header>
  );
}
