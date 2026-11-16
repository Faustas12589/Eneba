import { useEffect, useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "./Icons";

/**
 * A row that scrolls sideways, with ‹ › buttons on its sides like Eneba's
 * category icons. Each button only shows while there is more to see that way.
 */
export function ScrollRow({ className, label, children }: { className: string; label: string; children: React.ReactNode }) {
  const ref = useRef<HTMLElement>(null);
  const [canLeft, setCanLeft] = useState(false);
  const [canRight, setCanRight] = useState(false);

  const update = () => {
    const el = ref.current;
    if (!el) return;
    setCanLeft(el.scrollLeft > 0);
    setCanRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1); // -1: rounding on zoomed pages
  };

  useEffect(() => {
    update();
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, [children]);

  const scroll = (dir: 1 | -1) => ref.current?.scrollBy({ left: dir * ref.current.clientWidth * 0.8, behavior: "smooth" });

  return (
    <div className="scroll-row">
      {canLeft && (
        <button type="button" className="scroll-btn left" aria-label="Atgal" onClick={() => scroll(-1)}>
          <ChevronLeft size={28} />
        </button>
      )}
      <nav ref={ref} className={className} aria-label={label} onScroll={update}>
        {children}
      </nav>
      {canRight && (
        <button type="button" className="scroll-btn right" aria-label="Pirmyn" onClick={() => scroll(1)}>
          <ChevronRight size={28} />
        </button>
      )}
    </div>
  );
}
