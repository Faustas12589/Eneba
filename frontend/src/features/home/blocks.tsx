import { Link } from "react-router-dom";
import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { categoryLabel } from "../../lib/categories";
import { Status } from "../../components/Status";
import { ProductCard } from "../../components/ProductCard";
import { CategoryIcon } from "../../components/Icons";
import { InfoTip } from "../../components/InfoTip";
import { ScrollRow } from "../../components/ScrollRow";
import type { Product } from "../../types";
import { useAuth } from "../auth/AuthContext";

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

// What the ⓘ next to each title says: what the block shows and how it differs
// for a signed-in user and a guest.
const INFO = {
  recent:
    "Prekės, kurias neseniai atidarei, naujausios pirmos. Kai esi prisijungęs, sąrašas saugomas tavo paskyroje ir matomas bet kuriame įrenginyje; kitu atveju jis saugomas tik šioje naršyklėje. Prisijungus naršyklės istorija perkeliama į paskyrą.",
  recommended:
    "Rekomendacijos grindžiamos tavo peržiūrėtomis ir pirktomis prekėmis: siūlome panašaus žanro, kategorijos ir platformos prekes, kurių dar neturi. Kai esi prisijungęs, rekomendacijos remiasi tavo paskyros istorija, kitu atveju – šios naršyklės. Kol istorijos nėra, rodomos perkamiausios prekės.",
  bestSellers:
    "Daugiausiai kartų parduotos prekės visoje parduotuvėje. Sąrašas vienodas visiems pirkėjams, nesvarbu, ar esi prisijungęs.",
  giftCards:
    "Populiariausios el. dovanų kortelės pagal pardavimus. Kodą gausi iškart po apmokėjimo. Sąrašas tas pats visiems; į krepšelį dėti gali visi, o apmokėti – tik prisijungę pirkėjai.",
  topUps:
    "Populiariausi žaidimų platformų papildymai (Steam, PlayStation, Xbox ir kt.) pagal pardavimus. Sąrašas vienodas visiems pirkėjams; apmokant reikės prisijungti.",
  preOrders:
    "Dar neišleisti žaidimai. Užsisakyk dabar ir kodą gausi iškart; išleidimo dieną prekė savaime taps įprasta. Sąrašas vienodas visiems; apmokant reikės prisijungti.",
};

/** A titled row of product cards with an ⓘ; "Rodyti visus" is a button under the row. */
function ProductRail({ title, info, items, more }: { title: string; info: string; items: Product[]; more?: string }) {
  return (
    <>
      <div className="band-head">
        <h2>{title}</h2>
        <InfoTip text={info} />
      </div>
      <div className="grid rail">{items.map((p) => <ProductCard key={p.id} p={p} />)}</div>
      {more && (
        <div className="band-foot">
          <Link to={more} className="btn outline">Rodyti visus</Link>
        </div>
      )}
    </>
  );
}

/** A product block from one endpoint: loading/error inside its own band. */
function RailBlock({ title, info, load, more, personal }: { title: string; info: string; load: () => Promise<{ items: Product[] }>; more?: string; personal?: boolean }) {
  const uid = useAuth().user?.uid;
  const state = useAsync(load, personal ? [uid] : []); // personal blocks reload on sign-in/out
  return (
    <Band>
      <Status state={state}>{({ items }) => <ProductRail title={title} info={info} items={items} more={more} />}</Status>
    </Band>
  );
}

/** Full-bleed hero like Eneba's: the first banner is wide, the rest are narrow tall tiles. */
export function BannerBlock() {
  const state = useAsync(() => api.banners(), []);
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

/** "Neseniai peržiūrėta": no band at all until the browser has viewed something, like on Eneba. */
export function RecentlyViewedBlock() {
  const uid = useAuth().user?.uid;
  const state = useAsync(() => api.recentlyViewed(), [uid]);
  if (!state.data?.items.length) return null;
  return (
    <Band>
      <ProductRail title="Neseniai peržiūrėta" info={INFO.recent} items={state.data.items} />
    </Band>
  );
}

export const RecommendedBlock = () => <RailBlock title="Rekomenduojama tau" info={INFO.recommended} load={api.recommended} personal />;
export const BestSellersBlock = () => <RailBlock title="Perkamiausi" info={INFO.bestSellers} load={api.bestSellers} more="/store" />;
export const GiftCardsBlock = () => (
  <RailBlock title="Populiariausios el. dovanų kortelės" info={INFO.giftCards} load={api.giftCards} more="/store?category=gift-cards" />
);
export const TopUpsBlock = () => (
  <RailBlock title="Populiariausi papildymai" info={INFO.topUps} load={api.topUps} more="/store?category=gaming-cards" />
);
export const PreOrdersBlock = () => <RailBlock title="Įsigyk iš anksto" info={INFO.preOrders} load={api.preOrders} />;

// We have 5 categories; Eneba's row is full width. Repeating the list fills the
// row so the ‹ › arrows have something to scroll (the user asked for this look).
const MIN_ICONS = 16;

/** Eneba's icon row under the hero: icon + name, scrolls sideways with ‹ › buttons. */
export function CategoriesBlock() {
  const state = useAsync(() => api.categoryCounts(), []);
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
