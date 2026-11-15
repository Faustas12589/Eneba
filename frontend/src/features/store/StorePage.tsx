import { useSearchParams } from "react-router-dom";
import { api, type ProductQuery } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import { Pager } from "../../components/Pager";
import { formatItemCount } from "../../lib/format";

const SORTS = [
  { value: "", label: "Populiariausi" },
  { value: "price_asc", label: "Kaina: nuo mažiausios" },
  { value: "price_desc", label: "Kaina: nuo didžiausios" },
  { value: "discount", label: "Didžiausia nuolaida" },
  { value: "name", label: "Pavadinimas A–Ž" },
];

const list = (v: string | null) => (v ? v.split(",").filter(Boolean) : []);

/** All filter state lives in the URL, so results are shareable and back/forward works. */
export function StorePage() {
  const [sp, setSp] = useSearchParams();
  const query: ProductQuery = {
    q: sp.get("q") ?? "",
    platform: list(sp.get("platform")),
    region: list(sp.get("region")),
    genre: list(sp.get("genre")),
    minPrice: Number(sp.get("minPrice")) || undefined,
    maxPrice: Number(sp.get("maxPrice")) || undefined,
    sort: sp.get("sort") ?? "",
    page: Number(sp.get("page")) || 1,
  };
  const state = useAsync((signal) => api.products(query, signal), [sp.toString()]);

  function set(key: string, value: string) {
    const next = new URLSearchParams(sp);
    if (value) next.set(key, value);
    else next.delete(key);
    if (key !== "page") next.delete("page");
    setSp(next);
  }

  function toggle(key: "platform" | "region" | "genre", value: string) {
    const cur = list(sp.get(key));
    set(key, (cur.includes(value) ? cur.filter((v) => v !== value) : [...cur, value]).join(","));
  }

  // Price inputs are in euros for the user; the API takes cents.
  const euros = (cents?: number) => (cents ? String(cents / 100) : "");

  return (
    <div className="page container store">
      <aside className="filters">
        <Status state={state}>
          {({ facets }) => (
            <>
              {(
                [
                  ["platform", "Platforma", facets.platforms],
                  ["region", "Regionas", facets.regions],
                  ["genre", "Žanras", facets.genres],
                ] as const
              ).map(([key, label, values]) => (
                <fieldset key={key} style={{ border: 0, padding: 0, margin: 0 }}>
                  <label>{label}</label>
                  {values.map((v) => (
                    <label key={v} className="check">
                      <input type="checkbox" checked={list(sp.get(key)).includes(v)} onChange={() => toggle(key, v)} />
                      {v}
                    </label>
                  ))}
                </fieldset>
              ))}
            </>
          )}
        </Status>
        <div>
          <label>Kaina (EUR)</label>
          <input type="number" min={0} placeholder="Nuo" value={euros(query.minPrice)} onChange={(e) => set("minPrice", e.target.value ? String(Math.round(Number(e.target.value) * 100)) : "")} style={{ width: "48%" }} />{" "}
          <input type="number" min={0} placeholder="Iki" value={euros(query.maxPrice)} onChange={(e) => set("maxPrice", e.target.value ? String(Math.round(Number(e.target.value) * 100)) : "")} style={{ width: "48%" }} />
        </div>
      </aside>

      <section>
        <div className="toolbar">
          <h1>{query.q ? `Paieškos rezultatai: „${query.q}“` : "Parduotuvė"}</h1>
          <select value={query.sort} onChange={(e) => set("sort", e.target.value)} aria-label="Rūšiuoti">
            {SORTS.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
          </select>
        </div>
        <Status state={state}>
          {(pg) => (
            <>
              <p className="muted">Rasta: {formatItemCount(pg.total)}</p>
              {pg.items.length === 0 ? <p>Pagal šiuos filtrus prekių nerasta.</p> : (
                <div className="grid">{pg.items.map((p) => <ProductCard key={p.id} p={p} />)}</div>
              )}
              <Pager page={pg.page} pageSize={pg.pageSize} total={pg.total} onPage={(n) => set("page", String(n))} />
            </>
          )}
        </Status>
      </section>
    </div>
  );
}
