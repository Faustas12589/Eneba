// A visitor who has not signed in is this browser: a random id made once and
// kept in localStorage, sent as X-Guest-Id (DECISIONS #20). On sign-in the
// guest's views and orders move to the account (DECISIONS #28).
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
