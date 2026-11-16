import { useId } from "react";
import { InfoIcon } from "./Icons";

/**
 * The ⓘ next to a section title. Hovering or focusing it (keyboard, or a tap
 * on phones) shows a dark pop-up explaining the section. Pure CSS show/hide.
 */
export function InfoTip({ text }: { text: string }) {
  const id = useId();
  return (
    <span className="info-tip">
      <button type="button" className="info-btn" aria-label="Kas tai?" aria-describedby={id}>
        <InfoIcon size={22} />
      </button>
      <span role="tooltip" id={id} className="info-pop">{text}</span>
    </span>
  );
}
