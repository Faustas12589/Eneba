import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../../lib/api";
import { formatPrice } from "../../lib/format";
import type { Product } from "../../types";
import { SearchIcon } from "../../components/Icons";

/** Header search: debounced suggestions; Enter goes to /store?q=... (full results with filters). */
export function SearchBox() {
  const [q, setQ] = useState("");
  const [items, setItems] = useState<Product[]>([]);
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (q.trim().length < 2) {
      setItems([]);
      return;
    }
    const ctrl = new AbortController();
    const t = setTimeout(() => {
      api.suggest(q, ctrl.signal).then((r) => setItems(r.items)).catch(() => {});
    }, 200);
    return () => {
      clearTimeout(t);
      ctrl.abort();
    };
  }, [q]);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setOpen(false);
    navigate(`/store?q=${encodeURIComponent(q.trim())}`);
  }

  return (
    <form className="search" onSubmit={submit} onBlur={() => setTimeout(() => setOpen(false), 150)}>
      <span className="search-icon"><SearchIcon size={26} /></span>
      <input
        value={q}
        placeholder="Ieškoti žaidimų, papildymų ir daugiau"
        onChange={(e) => { setQ(e.target.value); setOpen(true); }}
        onFocus={() => setOpen(true)}
        aria-label="Paieška"
      />
      {open && items.length > 0 && (
        <div className="search-dropdown">
          {items.map((p) => (
            <Link key={p.id} to={`/product/${p.id}`} onClick={() => setOpen(false)}>
              <span>{p.name}</span>
              <span className="muted">{formatPrice(p.priceFrom, p.currency)}</span>
            </Link>
          ))}
        </div>
      )}
    </form>
  );
}
