import { createContext, useContext, useEffect, useState } from "react";
import {
  createUserWithEmailAndPassword,
  onAuthStateChanged,
  signInWithEmailAndPassword,
  signOut as fbSignOut,
} from "firebase/auth";
import { auth, firebaseConfigured, setDevSignedIn } from "../../lib/firebase";

/** The only user fields the app needs; works for Firebase and dev mode alike. */
export interface AppUser {
  uid: string;
  email: string | null;
}

interface AuthValue {
  user: AppUser | null;
  ready: boolean;
  /** false = dev mode: no Firebase, any email/password signs in as the dev user. */
  configured: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const Ctx = createContext<AuthValue | null>(null);
const DEV_KEY = "devUser";

function loadDevUser(): AppUser | null {
  try {
    const raw = sessionStorage.getItem(DEV_KEY);
    const u = raw ? (JSON.parse(raw) as AppUser) : null;
    setDevSignedIn(u !== null); // set before any child effect can call the API
    return u;
  } catch {
    return null;
  }
}

function saveDevUser(u: AppUser | null) {
  setDevSignedIn(u !== null);
  try {
    if (u) sessionStorage.setItem(DEV_KEY, JSON.stringify(u));
    else sessionStorage.removeItem(DEV_KEY);
  } catch {
    // storage blocked: dev login just won't survive a reload
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AppUser | null>(() => (auth ? null : loadDevUser()));
  const [ready, setReady] = useState(!auth);

  useEffect(() => {
    if (!auth) return;
    return onAuthStateChanged(auth, (u) => {
      setUser(u ? { uid: u.uid, email: u.email } : null);
      setReady(true);
    });
  }, []);

  // Dev mode only (no Firebase). Updates the token flag and state together.
  const setDevUser = (u: AppUser | null) => {
    saveDevUser(u);
    setUser(u);
  };
  const devSignIn = async (email: string) => setDevUser({ uid: "dev-user", email });

  const value: AuthValue = {
    user,
    ready,
    configured: firebaseConfigured,
    signIn: async (e, p) => (auth ? void (await signInWithEmailAndPassword(auth, e, p)) : devSignIn(e)),
    register: async (e, p) => (auth ? void (await createUserWithEmailAndPassword(auth, e, p)) : devSignIn(e)),
    signOut: async () => (auth ? fbSignOut(auth) : setDevUser(null)),
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useAuth(): AuthValue {
  const v = useContext(Ctx);
  if (!v) throw new Error("useAuth must be used inside <AuthProvider>");
  return v;
}
