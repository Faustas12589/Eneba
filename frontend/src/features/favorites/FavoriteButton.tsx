import { useEffect, useState } from "react";
import { HeartIcon } from "../../components/Icons";
import { useFavorites } from "./FavoritesContext";

/**
 * The heart. "icon" is the round button on a product card's cover; "full" is the
 * labelled button on the product page. aria-pressed tells screen readers the state.
 */
export function FavoriteButton({ productId, variant = "icon" }: { productId: string; variant?: "icon" | "full" }) {
  const { isFavorite, toggle } = useFavorites();
  const [error, setError] = useState<string | null>(null);
  const on = isFavorite(productId);

  useEffect(() => {
    if (!error) return;
    const t = setTimeout(() => setError(null), 3000);
    return () => clearTimeout(t);
  }, [error]);

  const label = on ? "Pašalinti iš norų sąrašo" : "Įtraukti į norų sąrašą";
  return (
    <span className={variant === "icon" ? "fav-wrap" : "fav-wrap fav-wrap-full"}>
      <button
        type="button"
        className={variant === "icon" ? "fav-toggle" : "fav-full"}
        aria-pressed={on}
        aria-label={label}
        title={label}
        onClick={async () => setError(await toggle(productId))}
      >
        <HeartIcon size={variant === "icon" ? 18 : 20} filled={on} />
        {variant === "full" && (on ? "Norų sąraše" : "Į norų sąrašą")}
      </button>
      {error && <span className="fav-error" role="alert">{error}</span>}
    </span>
  );
}
