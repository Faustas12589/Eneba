import { coverVariant } from "../lib/cover";
import { categoryLabel } from "../lib/categories";
import { CategoryIcon } from "./Icons";

interface CoverProps {
  productId: string;
  name: string;
  imageUrl?: string;
  category?: string; // order/cart lines don't carry it; then the cover has no icon or label
  size?: "card" | "large" | "thumb";
  className?: string; // layout class from the caller (card-img, cart-img)
}

/**
 * A product's cover. The seed has no cover art (Eneba's and publishers' images are
 * not ours: DECISIONS #24, #33), so when imageUrl is empty we draw one: a gradient
 * picked from the product id, the category icon as a watermark, the category and
 * the title. All colors come from tokens.css. A real imageUrl always wins.
 */
export function Cover({ productId, name, imageUrl, category, size = "card", className = "" }: CoverProps) {
  if (imageUrl) return <img className={className} src={imageUrl} alt={size === "thumb" ? "" : name} />;

  const cls = `${className} cover cover-${size} cover-v${coverVariant(productId)}`;
  if (size === "thumb") {
    // too small for text: the product's color and its first letter, like an app icon
    return <div className={cls} aria-hidden="true">{name.charAt(0)}</div>;
  }
  return (
    <div className={cls} role="img" aria-label={name}>
      {category && (
        <span className="cover-icon"><CategoryIcon id={category} size={size === "large" ? 160 : 96} /></span>
      )}
      <span className="cover-text">
        {category && <span className="cover-kicker">{categoryLabel(category)}</span>}
        <span className="cover-title">{name}</span>
      </span>
    </div>
  );
}
