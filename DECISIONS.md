# Decisions log

Append an entry whenever a non-obvious choice is made. Keep each to a few lines.

## 1. Layered backend instead of MVC
Go API returns JSON, so there is no "View". Layers: handler (HTTP) -> service (logic)
-> repository (Firestore, behind an interface). Why: testable without Firebase
(memory repo + fakes), and each layer has one reason to change.
Rejected: MVC (no real view layer); everything in handlers (untestable).

## 2. Filter / sort / search in Go, not in Firestore queries
Firestore cannot combine arbitrary filters + a sort without a composite index per
combination, and has no full-text or fuzzy search. For a catalog of up to a few thousand
products, the service caches the product list (60s TTL) and filters/sorts/searches in memory.
Trade-off: does not scale to millions of products; would move to Typesense/Algolia/Postgres then.

## 3. Typo-tolerant search via trigram similarity
`internal/search`: normalize -> trigram Jaccard + prefix/substring/word-prefix boosts.
Handles "eldne rign" and "witcher 3". Rejected: external search service (overkill, extra free-tier setup).

## 4. Memory repository for local dev/tests
`STORE=memory` serves embedded seed JSON so the app runs with zero Firebase setup,
and service tests run offline. Production uses `STORE=firestore`.

## 5. Money as integer cents
Avoids float rounding errors; formatting happens only in the UI.

## 6. Homepage blocks are independent
One endpoint and one React component per block, each with its own loading/error state,
so one failing block does not break the page.

## 7. Cart lives on the client (localStorage)
The cart holds offer ids + quantities in the browser, so guests can shop without
signing in and it costs no backend writes. The server re-reads prices and stock at
checkout, so a tampered cart cannot change what is charged.
Rejected: server-side cart in Firestore (needs auth before browsing, more writes, more code).

## 8. Simulated payment
Checkout skips a payment provider: "pay" succeeds immediately, then the order is
created and keys assigned. The assignment allows simulation; a provider adds setup
and webhooks without showing anything new about the marketplace itself.
Rejected: Stripe test mode (extra dependency + webhook handling for no graded value).

## 9. Discount % is derived, not stored
`service` computes `discountPercent` from `listPrice` and `priceFrom` (rounded, 0 if
the cheapest offer is not below list price). Seed data had badges that disagreed with
the prices (e.g. -5% on a product priced above list). One source of truth fixes that.
Rejected: trusting the stored field (can drift whenever an offer price changes).

## 10. Dev auth in memory mode
With `STORE=memory` the backend uses `middleware.DevVerifier`, which accepts only the
bearer token `dev` as one fixed user. When the frontend has no Firebase env vars, any
email/password "signs in" locally and sends that token. So the signed-in flow
(checkout, orders) runs offline and in tests. Safe because main.go wires DevVerifier
only in memory mode; production (`STORE=firestore`) always verifies real Firebase tokens.
Rejected: returning 503 without Firebase (checkout untestable offline);
a Firebase Auth emulator (extra tooling to install and explain).

## 11. Digital keys: seeded pool per offer (planned, checkout slice)
`seeds/keys.json` holds fake keys per offer; an offer's stock = its unsold keys.
Checkout runs one Firestore transaction: read the cart's offers, claim N unsold
keys (mark sold + orderId), write the order. Two buyers can't get the same key
because the transaction retries if a key changed under it. The memory repo does the
same under a mutex.
Rejected: generating a random key at purchase (no real stock to reserve, so the
"reserve stock" requirement would be a fake counter).

## 12. Firestore is backend-only; config via .env
Browsers never touch Firestore: the Go API uses the Admin SDK (service account),
which bypasses security rules, so `firestore.rules` denies all client access. One
place enforces auth and validation (the API), and the Firebase web config in the
frontend can't be abused to read or write data. Firebase Auth is still used client-side.
Local config comes from `backend/.env`, read by a ~20-line loader in `config`
(real env vars win, so Render is unaffected). Rejected: `godotenv` (a dependency for
20 lines); client-side Firestore SDK (rules become the security layer, logic splits in two).

<!-- 13. ... -->

## 13. Cart lines snapshot display data; logic is pure functions
Each cart line is one offer (not product) and stores a snapshot of name, seller,
price and stock next to offerId + quantity, so the cart page renders with no API call.
The snapshot is display-only: checkout re-reads price/stock on the server (#7).
Cart rules (add, clamp quantity to 1..stock, remove, totals) live in `features/cart/cart.ts`
as pure functions; `CartContext` only holds state and saves to localStorage.
Rejected: storing only ids and fetching each offer on the cart page (needs a new
batch endpoint and loading states for little gain); a reducer/state library (overkill).
