import { createContext, useContext, useEffect, useRef, useState } from "react";
import {
  createUserWithEmailAndPassword,
  GoogleAuthProvider,
  onAuthStateChanged,
  signInWithEmailAndPassword,
  signInWithPopup,
  signOut as fbSignOut,
  updateProfile,
  type User as FirebaseUser,
} from "firebase/auth";
import { auth, firebaseConfigured, setDevSignedIn } from "../../lib/firebase";
import { api } from "../../lib/api";

/** The user fields the app shows; works for Firebase and dev mode alike. */
export interface AppUser {
  uid: string;
  email: string | null;
  name: string | null;
  photoURL: string | null; // Google profile picture; null for email accounts
  provider: string; // "password", "google.com" or "dev"
  createdAt: string | null; // when the account was made (Firebase metadata)
}

interface AuthValue {
  user: AppUser | null;
  ready: boolean;
  /** false = dev mode: no Firebase, any email/password signs in as the dev user. */
  configured: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  signInWithGoogle: () => Promise<void>;
  updateName: (name: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const Ctx = createContext<AuthValue | null>(null);
const DEV_KEY = "devUser";

function toAppUser(u: FirebaseUser): AppUser {
  return {
    uid: u.uid,
    email: u.email,
    name: u.displayName,
    photoURL: u.photoURL,
    provider: u.providerData[0]?.providerId ?? "password",
    createdAt: u.metadata.creationTime ?? null,
  };
}

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

/**
 * Guests can do everything except pay (DECISIONS #28). When they sign in, their
 * guest views and orders move to the account first (`api.mergeGuest`), and only
 * then does `user` change, so pages that reload on a new user see the merged data.
 */
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AppUser | null>(() => (auth ? null : loadDevUser()));
  const [ready, setReady] = useState(!auth);
  // The sign-in in progress (Firebase sign-in + merge). The auth listener waits for it.
  const pending = useRef<Promise<void> | null>(null);

  useEffect(() => {
    if (!auth) return;
    return onAuthStateChanged(auth, async (u) => {
      await pending.current?.catch(() => {});
      setUser(u ? toAppUser(u) : null);
      setReady(true);
    });
  }, []);

  /** Runs a Firebase sign-in, then merges the guest data. Errors reach the form. */
  async function signInThenMerge(signIn: () => Promise<unknown>) {
    const p = (async () => {
      await signIn();
      // A failed merge must not block the sign-in; it is safe to retry next time.
      await api.mergeGuest().catch(() => {});
    })();
    pending.current = p;
    try {
      await p;
    } finally {
      if (pending.current === p) pending.current = null;
    }
  }

  // Dev mode only (no Firebase): one fake user, kept for this tab.
  async function devSignIn(email: string, name: string | null = null) {
    const u: AppUser = { uid: "dev-user", email, name, photoURL: null, provider: "dev", createdAt: new Date().toUTCString() };
    saveDevUser(u); // the dev token is sent from here on, so the merge is signed in
    await api.mergeGuest().catch(() => {});
    setUser(u);
  }

  const value: AuthValue = {
    user,
    ready,
    configured: firebaseConfigured,
    signIn: (email, password) =>
      auth ? signInThenMerge(() => signInWithEmailAndPassword(auth!, email, password)) : devSignIn(email),
    register: (email, password, name) =>
      auth
        ? signInThenMerge(async () => {
            const cred = await createUserWithEmailAndPassword(auth!, email, password);
            if (name) await updateProfile(cred.user, { displayName: name });
          })
        : devSignIn(email, name || null),
    signInWithGoogle: () =>
      auth ? signInThenMerge(() => signInWithPopup(auth!, new GoogleAuthProvider())) : devSignIn("dev@google"),
    updateName: async (name) => {
      if (!user) return;
      if (auth?.currentUser) {
        await updateProfile(auth.currentUser, { displayName: name });
        setUser(toAppUser(auth.currentUser)); // a new object, so React re-renders
      } else {
        const u = { ...user, name };
        saveDevUser(u);
        setUser(u);
      }
    },
    signOut: async () => {
      if (auth) await fbSignOut(auth);
      else {
        saveDevUser(null);
        setUser(null);
      }
    },
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useAuth(): AuthValue {
  const v = useContext(Ctx);
  if (!v) throw new Error("useAuth must be used inside <AuthProvider>");
  return v;
}
