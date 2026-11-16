import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { api, ApiError } from "../../lib/api";
import { useAuth } from "../auth/AuthContext";

interface FavoritesState {
  ids: string[]; // favorite product ids, newest first, as the server has them
  isFavorite: (productId: string) => boolean;
  /** Flips one product. Returns an error message to show, or null. */
  toggle: (productId: string) => Promise<string | null>;
}

const Ctx = createContext<FavoritesState | null>(null);

/**
 * The wishlist (DECISIONS #37). The server owns it (profiles/{userId}.favorites), for
 * guests too, like recently viewed; this only keeps the ids so every heart on the page
 * knows its state without its own request. Reloads when the user changes: on sign-in the
 * guest's favorites were already merged into the account (AuthContext runs the merge first).
 */
export function FavoritesProvider({ children }: { children: React.ReactNode }) {
  const { user, ready } = useAuth();
  const uid = user?.uid;
  const [ids, setIds] = useState<string[]>([]);

  useEffect(() => {
    if (!ready) return; // wait until we know who the user is, so hearts don't flicker
    let stale = false;
    api.favorites()
      .then((r) => !stale && setIds(r.items.map((p) => p.id)))
      .catch(() => !stale && setIds([]));
    return () => {
      stale = true;
    };
  }, [uid, ready]);

  // Optimistic: the heart changes at once; if the server says no, it flips back.
  const toggle = useCallback(async (productId: string) => {
    const on = !ids.includes(productId);
    const apply = (add: boolean) =>
      setIds((cur) => (add ? [productId, ...cur.filter((x) => x !== productId)] : cur.filter((x) => x !== productId)));
    apply(on);
    try {
      await api.setFavorite(productId, on);
      return null;
    } catch (e) {
      apply(!on);
      if (e instanceof ApiError && e.status === 409) return "Norų sąrašas pilnas: pašalink ką nors.";
      return "Nepavyko išsaugoti. Bandyk dar kartą.";
    }
  }, [ids]);

  const isFavorite = useCallback((productId: string) => ids.includes(productId), [ids]);

  return <Ctx.Provider value={{ ids, isFavorite, toggle }}>{children}</Ctx.Provider>;
}

export function useFavorites(): FavoritesState {
  const v = useContext(Ctx);
  if (!v) throw new Error("useFavorites must be used inside <FavoritesProvider>");
  return v;
}
