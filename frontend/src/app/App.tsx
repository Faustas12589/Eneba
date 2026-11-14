import { BrowserRouter, Outlet, Route, Routes } from "react-router-dom";
import { AuthProvider } from "../features/auth/AuthContext";
import { RequireAuth } from "../features/auth/RequireAuth";
import { LoginPage } from "../features/auth/LoginPage";
import { HomePage } from "../features/home/HomePage";
import { StorePage } from "../features/store/StorePage";
import { ProductPage } from "../features/product/ProductPage";
import { CartPage } from "../features/cart/CartPage";
import { CartProvider } from "../features/cart/CartContext";
import { CartPopup } from "../features/cart/CartPopup";
import { CheckoutPage } from "../features/checkout/CheckoutPage";
import { OrdersPage } from "../features/orders/OrdersPage";
import { Header } from "../components/Header";
import { Footer } from "../components/Footer";

function Layout() {
  return (
    <>
      <Header />
      <main><Outlet /></main>
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
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/store" element={<StorePage />} />
            <Route path="/product/:id" element={<ProductPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/checkout" element={<RequireAuth><CheckoutPage /></RequireAuth>} />
            <Route path="/orders" element={<RequireAuth><OrdersPage /></RequireAuth>} />
            <Route path="*" element={<div className="page container"><h1>Page not found</h1></div>} />
          </Route>
        </Routes>
        </CartProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
