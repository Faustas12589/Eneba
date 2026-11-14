import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { useAuth } from "../features/auth/AuthContext";
import { useCart } from "../features/cart/CartContext";

export function Header() {
  const { user, signOut } = useAuth();
  const { count } = useCart();
  return (
    <header className="header">
      <div className="container">
        <Link to="/" className="logo">ENEBA-STYLE</Link>
        <SearchBox />
        <nav className="nav">
          <Link to="/store">Store</Link>
          <Link to="/cart">Cart{count > 0 && <span className="cart-count">{count}</span>}</Link>
          {user ? (
            <>
              <Link to="/orders">Orders</Link>
              <button className="btn secondary" onClick={signOut}>Sign out</button>
            </>
          ) : (
            <Link to="/login" className="btn">Sign in</Link>
          )}
        </nav>
      </div>
    </header>
  );
}
