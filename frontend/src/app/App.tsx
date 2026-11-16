import { BrowserRouter, Outlet, Route, Routes } from "react-router-dom";
import { AuthProvider } from "../features/auth/AuthContext";
import { LoginPage } from "../features/auth/LoginPage";
import { AccountPage } from "../features/auth/AccountPage";
import { RequireAuth } from "../features/auth/RequireAuth";
import { HomePage } from "../features/home/HomePage";
import { StorePage } from "../features/store/StorePage";
import { ProductPage } from "../features/product/ProductPage";
import { CartPage } from "../features/cart/CartPage";
import { CartProvider } from "../features/cart/CartContext";
import { CartPopup } from "../features/cart/CartPopup";
import { CheckoutPage } from "../features/checkout/CheckoutPage";
import { OrdersPage } from "../features/orders/OrdersPage";
import { FavoritesProvider } from "../features/favorites/FavoritesContext";
import { FavoritesPage } from "../features/favorites/FavoritesPage";
import { Header } from "../components/Header";
import { Footer } from "../components/Footer";

function Layout() {
  return (
    <>
      <Header />
      <main className="main"><Outlet /></main>
      <Footer />
      <CartPopup />
    </>
  );
}

export function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <CartProvider>
        <FavoritesProvider>
        <Routes>
          {/* Like Eneba, the login page has no shop header/footer. */}
          <Route path="/login" element={<LoginPage />} />
          <Route element={<Layout />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/store" element={<StorePage />} />
            <Route path="/product/:id" element={<ProductPage />} />
            <Route path="/cart" element={<CartPage />} />
            {/* guests have a wishlist too, like recently viewed (DECISIONS #37) */}
            <Route path="/favorites" element={<FavoritesPage />} />
            {/* Guests see the checkout summary; pressing "Apmokėti" asks them to sign in (DECISIONS #28).
                Order history and the account page need an account (DECISIONS #32). */}
            <Route path="/checkout" element={<CheckoutPage />} />
            <Route path="/orders" element={<RequireAuth><OrdersPage /></RequireAuth>} />
            <Route path="/account" element={<RequireAuth><AccountPage /></RequireAuth>} />
            <Route path="*" element={<div className="page container"><h1>Puslapis nerastas</h1></div>} />
          </Route>
        </Routes>
        </FavoritesProvider>
        </CartProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
