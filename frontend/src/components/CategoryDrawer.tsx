import { useEffect } from "react";
import { Link } from "react-router-dom";
import { api } from "../lib/api";
import { useAsync } from "../lib/useAsync";
import { CATEGORIES } from "../lib/categories";
import { ChevronRight, CloseIcon } from "./Icons";

/**
 * Eneba's side menu behind "☰ Kategorijos": slides in from the left with every
 * product type, then every platform the store has (from the store's facets, so
 * a new platform shows up without a code change). Each row is a store filter.
 */
export function CategoryDrawer({ onClose }: { onClose: () => void }) {
  const platforms = useAsync((signal) => api.products({}, signal), []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="drawer-backdrop" onClick={onClose}>
      <aside className="drawer" aria-label="Kategorijos" onClick={(e) => e.stopPropagation()}>
        <div className="drawer-head">
          <Link to="/" className="logo" onClick={onClose}>eneba<span>style</span></Link>
          <button type="button" className="drawer-close" aria-label="Uždaryti" onClick={onClose}>
            <CloseIcon size={26} />
          </button>
        </div>
        {/* every link closes the drawer: following it is the whole point */}
        <nav className="drawer-list" onClick={(e) => (e.target as HTMLElement).closest("a") && onClose()}>
          <Link to="/store">Visi pasiūlymai</Link>
          {CATEGORIES.map((c) => (
            <Link key={c.id} to={`/store?category=${c.id}`}>
              {c.label} <ChevronRight size={18} />
            </Link>
          ))}
          {!!platforms.data?.facets.platforms.length && (
            <>
              <p className="drawer-section">Platformos</p>
              <div className="drawer-grid">
                {platforms.data.facets.platforms.map((p) => (
                  <Link key={p} to={`/store?platform=${encodeURIComponent(p)}`}>{p}</Link>
                ))}
              </div>
            </>
          )}
        </nav>
      </aside>
    </div>
  );
}
