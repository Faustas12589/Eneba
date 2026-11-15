// Login is switched off for now (DECISIONS #20). Orders belong to this browser
// instead: a random id made once and kept in localStorage, sent as X-Guest-Id.
const KEY = "guestId";
let fallback: string | null = null; // used when storage is blocked: lasts one page load

export function guestId(): string {
  try {
    let id = localStorage.getItem(KEY);
    if (!id) {
      id = crypto.randomUUID();
      localStorage.setItem(KEY, id);
    }
    return id;
  } catch {
    fallback ??= crypto.randomUUID();
    return fallback;
  }
}
