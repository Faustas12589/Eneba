import { BannerBlock, BestSellersBlock, DealsBlock } from "./blocks";

export function HomePage() {
  return (
    <div className="page container">
      <BannerBlock />
      <BestSellersBlock />
      <DealsBlock />
    </div>
  );
}
