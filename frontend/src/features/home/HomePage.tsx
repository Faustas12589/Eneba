import { api } from "../../lib/api";
import { useAsync } from "../../lib/useAsync";
import { Status } from "../../components/Status";
import type { HomeBlockRef } from "../../types";
import { BannerBlock, CategoriesBlock, ProductsBlock } from "./blocks";

/** Which component draws each block type the server can send. */
const DRAW: Record<HomeBlockRef["type"], (p: { blockRef: HomeBlockRef }) => React.ReactNode> = {
  banners: BannerBlock,
  categories: CategoriesBlock,
  products: ProductsBlock,
};

/**
 * Homepage = the blocks GET /api/home lists, in that order (DECISIONS #35).
 * Each block then loads its own endpoint. A type this build doesn't know is skipped,
 * so the server can add a new kind of block without breaking older frontends.
 */
export function HomePage() {
  const state = useAsync(() => api.homeLayout(), []);
  return (
    <div className="home">
      <Status state={state}>
        {({ blocks }) =>
          blocks.map((b) => {
            const Draw = DRAW[b.type];
            return Draw ? <Draw key={b.id} blockRef={b} /> : null;
          })
        }
      </Status>
    </div>
  );
}
