import { useState } from "react";
import { useAuth } from "./AuthContext";
import { GoogleIcon } from "../../components/Icons";

/** Firebase error codes -> text the shopper can act on. null = say nothing (they closed the popup). */
function authError(e: unknown): string | null {
  const code = (e as { code?: string } | null)?.code ?? "";
  switch (code) {
    case "auth/popup-closed-by-user":
    case "auth/cancelled-popup-request":
      return null;
    case "auth/invalid-credential":
    case "auth/wrong-password":
    case "auth/user-not-found":
      return "Neteisingas el. paštas arba slaptažodis.";
    case "auth/email-already-in-use":
      return "Šis el. paštas jau užregistruotas. Prisijunkite.";
    case "auth/weak-password":
      return "Slaptažodis per trumpas (min. 6 simboliai).";
    case "auth/invalid-email":
      return "Neteisingas el. pašto adresas.";
    case "auth/popup-blocked":
      return "Naršyklė užblokavo Google langą. Leiskite iššokančius langus ir bandykite dar kartą.";
    case "auth/too-many-requests":
      return "Per daug bandymų. Palaukite kelias minutes.";
    case "auth/unauthorized-domain":
    case "auth/operation-not-allowed":
      return "Šis prisijungimo būdas dar neįjungtas. Bandykite kitą.";
    default:
      return "Nepavyko prisijungti. Bandykite dar kartą.";
  }
}

type Mode = "login" | "register";

/**
 * Email/password (sign in or register) + Google, laid out like Eneba's login panel.
 * Used by /login and the checkout popup; `title` = false when the popup has its own heading.
 */
export function SignInForm({ onDone, initialMode = "login", title = true }: { onDone: () => void; initialMode?: Mode; title?: boolean }) {
  const { signIn, register, signInWithGoogle, configured } = useAuth();
  const [mode, setMode] = useState<Mode>(initialMode);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError(null);
    try {
      await action();
      onDone();
    } catch (e) {
      setError(authError(e));
    } finally {
      setBusy(false);
    }
  }

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    run(() => (mode === "login" ? signIn(email, password) : register(email, password, name.trim())));
  };

  const switchMode = () => {
    setMode(mode === "login" ? "register" : "login");
    setError(null);
  };

  return (
    <div className="signin">
      <div>
        {title && <h1 className="signin-title">{mode === "login" ? "Prisijungti" : "Susikurk paskyrą"}</h1>}
        <p className="signin-switch">
          {mode === "login" ? "Naujas vartotojas? " : "Jau turi paskyrą? "}
          <button type="button" onClick={switchMode}>{mode === "login" ? "Susikurk paskyrą" : "Prisijunk"}</button>
        </p>
      </div>
      {!configured && <p className="muted small">Kūrimo režimas: Firebase nesukonfigūruotas, todėl bet koks el. paštas ir slaptažodis prijungia vietinį testinį naudotoją.</p>}
      <button type="button" className="btn-google" disabled={busy} onClick={() => run(signInWithGoogle)}>
        <GoogleIcon size={22} />
        <span>Tęsti naudojant Google</span>
      </button>
      <p className="signin-or">arba</p>
      <form className="form" onSubmit={submit}>
        {mode === "register" && (
          <input placeholder="Vardas" autoComplete="name" value={name} onChange={(e) => setName(e.target.value)} />
        )}
        <input type="email" placeholder="El. paštas" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <input
          type="password"
          placeholder="Slaptažodis (min. 6 simboliai)"
          autoComplete={mode === "login" ? "current-password" : "new-password"}
          minLength={6}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
        {error && <p className="error" role="alert">{error}</p>}
        <button className="btn block signin-submit" disabled={busy}>
          {busy ? "Palaukite…" : mode === "login" ? "Prisijungti" : "Registruotis"}
        </button>
      </form>
    </div>
  );
}
