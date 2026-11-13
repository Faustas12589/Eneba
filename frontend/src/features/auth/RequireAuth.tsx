import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "./AuthContext";

/** Wrap routes that need a signed-in user (checkout, order history). */
export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, ready } = useAuth();
  const loc = useLocation();
  if (!ready) return <p className="muted">Loading…</p>;
  if (!user) return <Navigate to="/login" state={{ from: loc.pathname }} replace />;
  return <>{children}</>;
}
