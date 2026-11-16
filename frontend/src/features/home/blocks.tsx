import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { categoryLabel } from "../../lib/categories";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import { CategoryIcon } from "../../components/Icons";
import { InfoTip } from "../../components/InfoTip";
import { ScrollRow } from "../../components/ScrollRow";
import type { Banner, CategoryCount, HomeBlock, HomeBlockRef, Product } from "../../types";
import { useAuth } from "../auth/AuthContext";

// Each block fetches from its OWN endpoint and handles its own loading/error,
// so one failing service never blanks the whole homepage (DECISIONS #6).
// Titles, info texts and links come from that endpoint too: nothing about a
// block's content is written here, only how each type is drawn (DECISIONS #35).
// Every block is one full-width band; CSS alternates the band colors.

/** Loads one block from /api/home/{id}. Personal blocks reload on sign-in/out. */
function useBlock<T>(ref: HomeBlockRef) {
  const uid = useAuth().user?.uid;
  return useAsync(() => api.homeBlock<T>(ref.id, ref.personal), ref.personal ? [ref.id, uid] : [ref.id]);
}

function Band({ children }: { children: React.ReactNode }) {
  return (
    <section className="band">
      <div className="container">{children}</div>
    </section>
  );
}

/** A titled row of product cards with an ⓘ; "Rodyti visus" is a button under the row. */
function ProductRail({ block }: { block: HomeBlock<Product> }) {
  return (
    <>
      <div className="band-head">
        <h2>{block.title}</h2>
        {block.info && <InfoTip text={block.info} />}
      </div>
      <div className="grid rail">{block.items.map((p) => <ProductCard key={p.id} p={p} />)}</div>
      {block.moreUrl && (
        <div className="band-foot">
          <Link to={block.moreUrl} className="btn outline">Rodyti visus</Link>
        </div>
      )}
    </>
  );
}

/** A product rail block. With `hideWhenEmpty` (recently viewed) there is no band
 *  at all until it has items, like on Eneba. */
export function ProductsBlock({ blockRef }: { blockRef: HomeBlockRef }) {
  const state = useBlock<Product>(blockRef);
  if (state.data?.hideWhenEmpty && !state.data.items.length) return null;
  if (!state.data && blockRef.personal && state.loading) return null; // don't flash a band that may stay hidden
  return (
    <Band>
      <Status state={state}>{(block) => <ProductRail block={block} />}</Status>
    </Band>
  );
}

/** Full-bleed hero like Eneba's: the first banner is wide, the rest are narrow tall tiles. */
export function BannerBlock({ blockRef }: { blockRef: HomeBlockRef }) {
  const state = useBlock<Banner>(blockRef);
  return (
    <section className="hero">
      <Status state={state}>
        {({ items }) =>
          items.map((b) => (
            <Link key={b.id} to={b.linkUrl} className="banner" style={b.imageUrl ? { backgroundImage: `url(${b.imageUrl})`, backgroundSize: "cover" } : undefined}>
              <h2>{b.title}</h2>
              <p>{b.subtitle}</p>
            </Link>
          ))
        }
      </Status>
    </section>
  );
}

// We have 5 categories; Eneba's row is full width. Repeating the list fills the
// row so the ‹ › arrows have something to scroll (the user asked for this look).
const MIN_ICONS = 16;

/** Eneba's icon row under the hero: icon + name, scrolls sideways with ‹ › buttons. */
export function CategoriesBlock({ blockRef }: { blockRef: HomeBlockRef }) {
  const state = useBlock<CategoryCount>(blockRef);
  return (
    <section className="category-strip">
      <div className="container">
        <Status state={state}>
          {({ items }) => {
            const filled = items.length ? Array.from({ length: Math.max(MIN_ICONS, items.length) }, (_, i) => items[i % items.length]) : [];
            return (
              <ScrollRow className="category-icons" label="Kategorijos">
                {filled.map((c, i) => (
                  <Link key={`${c.category}-${i}`} to={`/store?category=${c.category}`} className="category-icon">
                    <CategoryIcon id={c.category} size={32} />
                    <span className="category-icon-name">{categoryLabel(c.category)}</span>
                  </Link>
                ))}
              </ScrollRow>
            );
          }}
        </Status>
      </div>
    </section>
  );
}
