import {
  BannerBlock,
  BestSellersBlock,
  CategoriesBlock,
  GiftCardsBlock,
  PreOrdersBlock,
  RecentlyViewedBlock,
  RecommendedBlock,
  TopUpsBlock,
} from "./blocks";

/** Homepage = a stack of independent blocks; each is a full-width band, colors alternate (CSS). */
export function HomePage() {
  return (
    <div className="home">
      <BannerBlock />
      <CategoriesBlock />
      <RecentlyViewedBlock />
      <RecommendedBlock />
      <BestSellersBlock />
      <GiftCardsBlock />
      <TopUpsBlock />
      <PreOrdersBlock />
    </div>
  );
}
