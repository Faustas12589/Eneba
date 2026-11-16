import { useState } from "react";
import { UserIcon } from "../../components/Icons";
import type { AppUser } from "./AuthContext";

/**
 * The profile picture (from Google), or the plain user icon when there is none
 * or it fails to load. No initials, on purpose.
 */
export function Avatar({ user, size = 40 }: { user: AppUser | null; size?: number }) {
  const [broken, setBroken] = useState<string | null>(null); // the photo URL that failed
  const photo = user?.photoURL && user.photoURL !== broken ? user.photoURL : null;
  return (
    <span className="avatar" style={{ width: size, height: size }}>
      {photo ? (
        // no-referrer: Google's image server can refuse requests that carry a referrer
        <img src={photo} alt="" referrerPolicy="no-referrer" onError={() => setBroken(photo)} />
      ) : (
        <UserIcon size={Math.round(size * 0.55)} />
      )}
    </span>
  );
}
