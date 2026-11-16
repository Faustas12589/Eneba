import { useEffect } from "react";
import { SignInForm } from "./SignInForm";

/** "Prisijungimas būtinas" popup: shown when a guest presses pay at checkout. */
export function SignInModal({ onClose, onSignedIn }: { onClose: () => void; onSignedIn: () => void }) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" role="dialog" aria-modal="true" aria-labelledby="signin-title" onClick={(e) => e.stopPropagation()}>
        <button type="button" className="modal-close" aria-label="Uždaryti" onClick={onClose}>×</button>
        <h2 id="signin-title">Prisijungimas būtinas</h2>
        <p className="muted small">Norėdami apmokėti, prisijunkite arba susikurkite paskyrą. Jūsų krepšelis išliks.</p>
        <SignInForm onDone={onSignedIn} title={false} />
      </div>
    </div>
  );
}
