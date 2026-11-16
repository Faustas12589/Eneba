import { useEffect, useState } from "react";

interface Props {
  /** Cheapest and dearest product for the other filters (cents, from the API's facets). */
  bounds: { min: number; max: number };
  /** The price filter in the URL (cents); undefined = no limit on that side. */
  from?: number;
  to?: number;
  /** Save the new range; undefined for a side that is back at its bound (= no limit). */
  onApply: (from?: number, to?: number) => void;
}

// The slider works in whole euros; the URL and the API stay in cents (rule 2).
const floorEur = (cents: number) => Math.floor(cents / 100);
const ceilEur = (cents: number) => Math.ceil(cents / 100);

/**
 * "Kaina" filter: a two-handle slider plus "Nuo"/"Iki" boxes in euros. Dragging
 * only changes what is shown; the store reloads when a handle is let go or a box
 * is left/Entered, so typing "25" does not fire a request for "2" first.
 */
export function PriceFilter({ bounds, from, to, onApply }: Props) {
  const lo = floorEur(bounds.min);
  const hi = Math.max(ceilEur(bounds.max), lo + 1); // never a zero-width slider
  const [range, setRange] = useState<[number, number]>([lo, hi]);

  // Follow the URL and the bounds (another filter changed, or back/forward).
  useEffect(() => {
    setRange([from ? floorEur(from) : lo, to ? ceilEur(to) : hi]);
  }, [from, to, lo, hi]);

  const [a, b] = range; // NaN while a box is empty
  // A box left empty means "no limit" on that side, i.e. the bound.
  const clamp = (v: number, empty: number) => (Number.isNaN(v) ? empty : Math.min(hi, Math.max(lo, Math.round(v))));
  const sliderA = Number.isNaN(a) ? lo : a;
  const sliderB = Number.isNaN(b) ? hi : b;

  function apply(next: [number, number] = [sliderA, sliderB]) {
    const [x, y] = [Math.min(next[0], next[1]), Math.max(next[0], next[1])];
    setRange([x, y]);
    const nextFrom = x > lo ? x * 100 : undefined;
    const nextTo = y < hi ? y * 100 : undefined;
    if (nextFrom !== from || nextTo !== to) onApply(nextFrom, nextTo); // e.g. a box left unchanged: no reload
  }

  const pct = (v: number) => ((v - lo) / (hi - lo)) * 100;
  const track = { "--from": `${pct(sliderA)}%`, "--to": `${pct(sliderB)}%` } as React.CSSProperties;

  return (
    <fieldset className="price-filter">
      <legend>Kaina</legend>
      <div className="price-slider" style={track}>
        <input
          type="range" min={lo} max={hi} value={sliderA} aria-label="Mažiausia kaina"
          onChange={(e) => setRange([Math.min(Number(e.target.value), sliderB), sliderB])}
          onPointerUp={() => apply()} onKeyUp={() => apply()}
        />
        <input
          type="range" min={lo} max={hi} value={sliderB} aria-label="Didžiausia kaina"
          onChange={(e) => setRange([sliderA, Math.max(Number(e.target.value), sliderA)])}
          onPointerUp={() => apply()} onKeyUp={() => apply()}
        />
      </div>
      <div className="price-boxes">
        <label>
          <span>Nuo</span>
          <input
            type="number" min={lo} max={hi} value={Number.isNaN(a) ? "" : a} placeholder={String(lo)}
            onChange={(e) => setRange([e.target.valueAsNumber, b])}
            onBlur={() => apply([clamp(a, lo), sliderB])}
            onKeyDown={(e) => e.key === "Enter" && apply([clamp(a, lo), sliderB])}
          />
          <span className="muted">€</span>
        </label>
        <span className="muted">–</span>
        <label>
          <span>Iki</span>
          <input
            type="number" min={lo} max={hi} value={Number.isNaN(b) ? "" : b} placeholder={String(hi)}
            onChange={(e) => setRange([a, e.target.valueAsNumber])}
            onBlur={() => apply([sliderA, clamp(b, hi)])}
            onKeyDown={(e) => e.key === "Enter" && apply([sliderA, clamp(b, hi)])}
          />
          <span className="muted">€</span>
        </label>
      </div>
    </fieldset>
  );
}
