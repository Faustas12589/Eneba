import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { categoryLabel } from "../../lib/categories";
import { formatItemCount } from "../../lib/format";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import type { Product } from "../../types";

// Each block fetches from its OWN endpoint and handles its own loading/error,
// so one failing service never blanks the whole homepage (DECISIONS #6).
// Every block is one full-width band; CSS alternates the band colors.

function Band({ children }: { children: React.ReactNode }) {
  return (
    <section className="band">
      <div className="container">{children}</div>
    </section>
  );
}

/** A titled row of product cards; "Rodyti visus" links to the matching store page. */
function ProductRail({ title, items, more }: { title: string; items: Product[]; more?: string }) {
  return (
    <>
      <div className="band-head">
        <h2>{title}</h2>
        {more && <Link to={more} className="band-more">Rodyti visus</Link>}
      </div>
      <div className="grid rail">{items.map((p) => <ProductCard key={p.id} p={p} />)}</div>
    </>
  );
}

/** A product block from one endpoint: loading/error inside its own band. */
function RailBlock({ title, load, more }: { title: string; load: () => Promise<{ items: Product[] }>; more?: string }) {
  const state = useAsync(load, []);
  return (
    <Band>
      <Status state={state}>{({ items }) => <ProductRail title={title} items={items} more={more} />}</Status>
    </Band>
  );
}

export function BannerBlock() {
  const state = useAsync(() => api.banners(), []);
  return (
    <Band>
      <Status state={state}>
        {({ items }) => (
          <div className="banner-row">
            {items.map((b) => (
              <Link key={b.id} to={b.linkUrl} className="banner" style={b.imageUrl ? { backgroundImage: `url(${b.imageUrl})`, backgroundSize: "cover" } : undefined}>
                <h2>{b.title}</h2>
                <p>{b.subtitle}</p>
              </Link>
            ))}
          </div>
        )}
      </Status>
    </Band>
  );
}

/** "Neseniai peržiūrėta": no band at all until the browser has viewed something, like on Eneba. */
export function RecentlyViewedBlock() {
  const state = useAsync(() => api.recentlyViewed(), []);
  if (!state.data?.items.length) return null;
  return (
    <Band>
      <ProductRail title="Neseniai peržiūrėta" items={state.data.items} />
    </Band>
  );
}

export const RecommendedBlock = () => <RailBlock title="Rekomenduojama tau" load={api.recommended} />;
export const BestSellersBlock = () => <RailBlock title="Perkamiausi" load={api.bestSellers} more="/store" />;
export const GiftCardsBlock = () => (
  <RailBlock title="Populiariausios el. dovanų kortelės" load={api.giftCards} more="/store?category=gift-cards" />
);
export const TopUpsBlock = () => (
  <RailBlock title="Populiariausi papildymai" load={api.topUps} more="/store?category=gaming-cards" />
);
export const PreOrdersBlock = () => <RailBlock title="Įsigyk iš anksto" load={api.preOrders} />;

/** Category tiles with how many products each has; a tile opens the filtered store. */
export function CategoriesBlock() {
  const state = useAsync(() => api.categoryCounts(), []);
  return (
    <Band>
      <h2>Kategorijos</h2>
      <Status state={state}>
        {({ items }) => (
          <div className="category-tiles">
            {items.map((c) => (
              <Link key={c.category} to={`/store?category=${c.category}`} className="category-tile">
                <span className="category-tile-name">{categoryLabel(c.category)}</span>
                <span className="muted small">{formatItemCount(c.count)}</span>
              </Link>
            ))}
          </div>
        )}
      </Status>
    </Band>
  );
}
