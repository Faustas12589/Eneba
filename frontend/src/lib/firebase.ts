import { initializeApp, type FirebaseApp } from "firebase/app";
import { getAuth, type Auth } from "firebase/auth";

const cfg = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY as string | undefined,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN as string | undefined,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID as string | undefined,
  appId: import.meta.env.VITE_FIREBASE_APP_ID as string | undefined,
};

export const firebaseConfigured = Boolean(cfg.apiKey && cfg.projectId);

let app: FirebaseApp | null = null;
export let auth: Auth | null = null;

if (firebaseConfigured) {
  app = initializeApp(cfg);
  auth = getAuth(app);
}

// Dev mode (no Firebase env vars): the backend in STORE=memory accepts this token.
// See DECISIONS #10. Set by AuthContext when the dev user signs in/out.
const DEV_TOKEN = "dev";
let devSignedIn = false;
export function setDevSignedIn(v: boolean) {
  devSignedIn = v;
}

export async function getIdToken(): Promise<string | null> {
  if (auth) return auth.currentUser ? auth.currentUser.getIdToken() : null;
  return devSignedIn ? DEV_TOKEN : null;
}
