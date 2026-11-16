import { useSearchParams } from "react-router-dom";
import { api, type ProductQuery } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import { Pager } from "../../components/Pager";
import { formatItemCount, regionLabel } from "../../lib/format";
import { categoryLabel } from "../../lib/categories";
import { PriceFilter } from "./PriceFilter";

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
    category: list(sp.get("category")),
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

  // Both ends of the price range change in one URL update (cents, like the API).
  function setPrice(from?: number, to?: number) {
    const next = new URLSearchParams(sp);
    for (const [key, v] of [["minPrice", from], ["maxPrice", to]] as const) {
      if (v) next.set(key, String(v));
      else next.delete(key);
    }
    next.delete("page");
    setSp(next);
  }

  return (
    <div className="page container store">
      <aside className="filters">
        <Status state={state}>
          {({ facets }) => (
            <>
              {/* No "Kategorija" checkboxes: the category comes from the header row or the
                  side menu. The price range comes first instead. */}
              <PriceFilter
                bounds={{ min: facets.priceMin, max: facets.priceMax }}
                from={query.minPrice}
                to={query.maxPrice}
                onApply={setPrice}
              />
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
                      {key === "region" ? regionLabel(v) : v}
                    </label>
                  ))}
                </fieldset>
              ))}
            </>
          )}
        </Status>
      </aside>

      <section>
        <div className="toolbar">
          <h1>
            {query.q
              ? `Paieškos rezultatai: „${query.q}“`
              : query.category?.length === 1
                ? categoryLabel(query.category[0])
                : "Parduotuvė"}
          </h1>
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
