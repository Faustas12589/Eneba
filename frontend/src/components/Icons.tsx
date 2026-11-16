/**
 * Small line icons drawn with `currentColor`, so their color comes from the
 * surrounding text (and so from tokens.css). Hand-written SVG: no icon library.
 */
type P = { size?: number };

function Svg({ size = 24, children }: P & { children: React.ReactNode }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor"
      strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {children}
    </svg>
  );
}

export const CartIcon = (p: P) => (
  <Svg {...p}><circle cx="9" cy="20" r="1.4" /><circle cx="18" cy="20" r="1.4" /><path d="M2 3h3l2.4 12h11.2L21 7H6.2" /></Svg>
);
export const UserIcon = (p: P) => (
  <Svg {...p}><circle cx="12" cy="9" r="3.5" /><path d="M5 20c1.2-3.5 4-5 7-5s5.8 1.5 7 5" /></Svg>
);
export const MenuIcon = (p: P) => <Svg {...p}><path d="M3 6h18M3 12h18M3 18h18" /></Svg>;
export const SearchIcon = (p: P) => <Svg {...p}><circle cx="11" cy="11" r="7" /><path d="M20 20l-4-4" /></Svg>;
export const ChevronLeft = (p: P) => <Svg {...p}><path d="M15 5l-7 7 7 7" /></Svg>;
export const ChevronRight = (p: P) => <Svg {...p}><path d="M9 5l7 7-7 7" /></Svg>;
export const CloseIcon = (p: P) => <Svg {...p}><path d="M5 5l14 14M19 5L5 19" /></Svg>;
/** Heart for favorites; `filled` when the product is on the wishlist. */
export function HeartIcon({ size = 24, filled = false }: P & { filled?: boolean }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill={filled ? "currentColor" : "none"} stroke="currentColor"
      strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M12 20s-7-4.4-9-9.2C1.7 7.6 3.6 4.5 6.9 4.5c2 0 3.3 1.1 5.1 3 1.8-1.9 3.1-3 5.1-3 3.3 0 5.2 3.1 3.9 6.3-2 4.8-9 9.2-9 9.2z" />
    </svg>
  );
}
/** Three sliders: the "Filtrai" button on phones. */
export const FilterIcon = (p: P) => (
  <Svg {...p}><path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12" /><circle cx="16" cy="6" r="2" /><circle cx="10" cy="12" r="2" /><circle cx="18" cy="18" r="2" /></Svg>
);
export const InfoIcon = (p: P) => (
  <Svg {...p}><circle cx="12" cy="12" r="9" /><path d="M12 11v6" /><circle cx="12" cy="7.5" r="0.6" /></Svg>
);

/** Google's four-color "G" for the sign-in button (their branding asks for it). Colors are tokens. */
export function GoogleIcon({ size = 20 }: P) {
  return (
    <svg width={size} height={size} viewBox="0 0 18 18" aria-hidden="true">
      <path style={{ fill: "var(--color-google-blue)" }} d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.91c1.7-1.57 2.69-3.87 2.69-6.62z" />
      <path style={{ fill: "var(--color-google-green)" }} d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.91-2.26c-.81.54-1.84.86-3.05.86-2.34 0-4.33-1.58-5.04-3.71H.96v2.33A9 9 0 0 0 9 18z" />
      <path style={{ fill: "var(--color-google-yellow)" }} d="M3.96 10.71A5.4 5.4 0 0 1 3.68 9c0-.59.1-1.17.28-1.71V4.96H.96A9 9 0 0 0 0 9c0 1.45.35 2.83.96 4.04l3-2.33z" />
      <path style={{ fill: "var(--color-google-red)" }} d="M9 3.58c1.32 0 2.51.45 3.44 1.35l2.58-2.58A9 9 0 0 0 .96 4.96l3 2.33C4.67 5.16 6.66 3.58 9 3.58z" />
    </svg>
  );
}

/** One icon per category id (domain.Categories). */
export function CategoryIcon({ id, size = 36 }: P & { id: string }) {
  switch (id) {
    case "games": // gamepad
      return <Svg size={size}><rect x="2" y="7" width="20" height="11" rx="5" /><path d="M7 11v3M5.5 12.5h3" /><circle cx="16" cy="11.5" r=".8" /><circle cx="18" cy="13.5" r=".8" /></Svg>;
    case "dlc": // puzzle piece = add-on
      return <Svg size={size}><path d="M5 8h3a2 2 0 1 1 4 0h3v3a2 2 0 1 1 0 4v4H5v-4a2 2 0 1 0 0-4z" /></Svg>;
    case "gaming-cards": // card with a gamepad
      return <Svg size={size}><rect x="2" y="5" width="20" height="14" rx="2" /><rect x="7" y="10" width="10" height="5" rx="2.5" /></Svg>;
    case "gift-cards": // gift box
      return <Svg size={size}><rect x="3" y="9" width="18" height="11" rx="1" /><path d="M3 13h18M12 9v11M12 9c-2-4-6-4-6-1.5S10 9 12 9c2 0 6 0 6-1.5S14 5 12 9" /></Svg>;
    default: // software: monitor
      return <Svg size={size}><rect x="3" y="4" width="18" height="12" rx="1.5" /><path d="M8 20h8M12 16v4" /></Svg>;
  }
}
