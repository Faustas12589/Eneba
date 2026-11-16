import { useLocation, useNavigate } from "react-router-dom";
import { SignInForm } from "./SignInForm";

/** /login: the same form as the checkout popup, then back to where the user came from. */
export function LoginPage() {
  const navigate = useNavigate();
  const from = (useLocation().state as { from?: string } | null)?.from ?? "/";
  return (
    <div className="page container narrow-sm">
      <h1>Prisijungti</h1>
      <SignInForm onDone={() => navigate(from, { replace: true })} />
    </div>
  );
}
