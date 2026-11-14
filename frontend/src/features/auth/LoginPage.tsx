import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";

export function LoginPage() {
  const { signIn, register, configured } = useAuth();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const navigate = useNavigate();
  const from = (useLocation().state as { from?: string } | null)?.from ?? "/";

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await (mode === "login" ? signIn : register)(email, password);
      navigate(from, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kažkas nepavyko");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="page container">
      <h1>{mode === "login" ? "Prisijungti" : "Sukurti paskyrą"}</h1>
      {!configured && <p className="muted">Kūrimo režimas: Firebase nesukonfigūruotas, todėl bet koks el. paštas ir slaptažodis prijungia vietinį testinį naudotoją.</p>}
      <form className="form" onSubmit={submit}>
        <input type="email" placeholder="El. paštas" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <input type="password" placeholder="Slaptažodis (min. 6 simboliai)" minLength={6} value={password} onChange={(e) => setPassword(e.target.value)} required />
        {error && <p className="error">{error}</p>}
        <button className="btn" disabled={busy}>{mode === "login" ? "Prisijungti" : "Registruotis"}</button>
        <button type="button" className="btn secondary" onClick={() => setMode(mode === "login" ? "register" : "login")}>
          {mode === "login" ? "Neturite paskyros? Registruokitės" : "Jau turite paskyrą? Prisijunkite"}
        </button>
      </form>
    </div>
  );
}
