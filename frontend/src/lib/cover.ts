/** How many cover gradients tokens.css defines (--cover-1-* .. --cover-6-*). */
export const COVER_VARIANTS = 6;

/**
 * Picks a cover gradient (1..COVER_VARIANTS) from a product id. The same id
 * always gives the same number, so a product has one color on its card, its
 * product page, in the cart and in orders. A plain string hash: no randomness,
 * nothing stored.
 */
export function coverVariant(id: string): number {
  let h = 0;
  for (const ch of id) h = (h * 31 + ch.charCodeAt(0)) >>> 0; // >>> 0 keeps it an unsigned 32-bit int
  return (h % COVER_VARIANTS) + 1;
}
