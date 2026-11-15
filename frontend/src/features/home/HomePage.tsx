import { BannerBlock, BestSellersBlock, DealsBlock, RecentlyViewedBlock, RecommendedBlock } from "./blocks";

export function HomePage() {
  return (
    <div className="page container">
      <BannerBlock />
      <RecentlyViewedBlock />
      <RecommendedBlock />


      <BestSellersBlock />
                  {/*"populiariausios el. dovanų kortelės"*/}
            {/*"populiariausi papildymai"(roblux steam playstation.store)*/}
            {/*"įsigyk iš anksto"*/}

      <DealsBlock /> {/* this one isnt needed */}
      {/* game categories with the number of the games that each category has */}
    </div>
  );
}

{/* each of the block need to be in the alternating color from purple to dark purple (the ones you've used)*/}