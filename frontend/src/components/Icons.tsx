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
