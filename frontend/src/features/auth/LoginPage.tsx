import { Link, useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { SignInForm } from "./SignInForm";

/**
 * /login (and /login?mode=register): a page of its own like Eneba's, without the
 * shop header: logo, a big greeting on the left, the sign-in panel on the right.
 * After signing in, the user goes back to where they came from.
 */
export function LoginPage() {
  const navigate = useNavigate();
  const from = (useLocation().state as { from?: string } | null)?.from ?? "/";
  const [sp] = useSearchParams();
  const mode = sp.get("mode") === "register" ? "register" : "login";

  return (
    <div className="login-page">
      <header className="login-top">
        <Link to="/" className="logo">eneba<span>style</span></Link>
      </header>
      <main className="login-main">
        <p className="login-hello">Sveiki!<br />Malonu jus matyti!</p>
        {/* key: switching Prisijungti/Registruotis in the header resets the form */}
        <section className="login-panel">
          <SignInForm key={mode} initialMode={mode} onDone={() => navigate(from, { replace: true })} />
        </section>
      </main>
      <footer className="login-foot">
        <Link to="/">Į parduotuvę</Link>
        <span>Autorinės teisės © {new Date().getFullYear()} enebastyle. Mokomasis projektas.</span>
      </footer>
    </div>
  );
}
