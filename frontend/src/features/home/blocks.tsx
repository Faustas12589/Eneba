import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import type { Product } from "../../types";

// Each block fetches from its OWN endpoint and handles its own loading/error,
// so one failing service never blanks the whole homepage.

export function BannerBlock() {
  const state = useAsync(() => api.banners(), []);
  return (
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
  );
}

function ProductRail({ title, items }: { title: string; items: Product[] }) {
  return (
    <section className="block">
      <h2>{title}</h2>
      <div className="grid rail">{items.map((p) => <ProductCard key={p.id} p={p} />)}</div>
    </section>
  );
}

export function BestSellersBlock() {
  const state = useAsync(() => api.bestSellers(), []);
  return <Status state={state}>{({ items }) => <ProductRail title="Perkamiausi" items={items} />}</Status>;
}

export function DealsBlock() {
  const state = useAsync(() => api.deals(), []);
  return <Status state={state}>{({ items }) => <ProductRail title="Geriausi pasiūlymai" items={items} />}</Status>;
}
