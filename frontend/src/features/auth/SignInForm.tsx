import { useState } from "react";
import { useAuth } from "./AuthContext";

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

/** Email/password (sign in or register) + Google. Used by /login and the checkout popup. */
export function SignInForm({ onDone }: { onDone: () => void }) {
  const { signIn, register, signInWithGoogle, configured } = useAuth();
  const [mode, setMode] = useState<"login" | "register">("login");
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

  return (
    <div className="signin">
      {!configured && <p className="muted small">Kūrimo režimas: Firebase nesukonfigūruotas, todėl bet koks el. paštas ir slaptažodis prijungia vietinį testinį naudotoją.</p>}
      <button type="button" className="btn secondary block" disabled={busy} onClick={() => run(signInWithGoogle)}>
        Tęsti su Google
      </button>
      <p className="signin-or muted small">arba</p>
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
        <button className="btn block" disabled={busy}>
          {busy ? "Palaukite…" : mode === "login" ? "Prisijungti" : "Registruotis"}
        </button>
      </form>
      <button type="button" className="link-btn" onClick={() => { setMode(mode === "login" ? "register" : "login"); setError(null); }}>
        {mode === "login" ? "Neturite paskyros? Registruokitės" : "Jau turite paskyrą? Prisijunkite"}
      </button>
    </div>
  );
}
