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

## 14. Live stock check on add-to-cart; popup + header dropdown
Before adding an offer (product page button or "+" in the cart), the frontend re-reads
that offer's stock from `GET /api/products/{id}`. If the cart already holds all of it,
or someone else bought it since the page loaded, the user gets an error instead of a
silent no-op. Success opens an Eneba-style popup ("Pridėta į krepšelį"); the header
cart is a dropdown with the same `CartLine`/`QtyStepper` components as the cart page.
This is UX only: it can't reserve stock, so checkout still re-checks in a transaction (#11).
Rejected: trusting the stock loaded with the page (misses other buyers); a dedicated
`GET /api/offers/{id}` endpoint (the product endpoint already returns it).
Note: cart/popup text is Lithuanian to match eneba.lt; the rest of the UI is still English.

## 15. Stock is reserved when an item goes into the cart (supersedes the stock part of #7, #14)
`PUT /api/carts/{cartId}/items/{offerId} {"quantity": n}` makes a cart hold exactly n
units: one atomic step (Firestore transaction / mutex in memory) moves the difference
between `offers.stock` and a `reservations/{cartId}_{offerId}` doc. So `offer.stock` is
what anyone else can still buy, and two carts can never get the same last unit (test:
`TestSetQuantityNoOversell`). The rule (`reserve`) is a pure function in `service`; the
repository only runs it atomically (`UpdateReservation(..., apply)`), keeping layering.
- Absolute quantity, not "+1": a retried request can't reserve twice.
- A cart can only release what its own reservation holds, so nobody can inflate stock.
- `cartId` is a browser-generated UUID (guest carts still work, no login needed);
  max 10 per line so one cart can't hoard an offer.
- 409 returns the real stock + reserved qty; the frontend corrects its cart from it.
Known gap: abandoned carts hold stock forever. Fix later: expire reservations after
~30 min (`updatedAt` is stored for that) or convert them at checkout.
Rejected: reserving only at checkout (user asked for cart-time reservation);
decrementing without a reservation record (release couldn't be verified).

## 16. Cart reservations expire after 30 minutes
Closes the gap in #15: an abandoned cart would otherwise hold stock forever. Each line
is held for 30 min after it was last changed (`reservations.updatedAt`). A goroutine in
`cmd/api` calls `Cart.ReleaseExpired` every minute: it lists old reservations, then
releases each in its own transaction that re-checks the timestamp, so a line the
shopper touched a second ago is left alone. Several API instances can sweep safely.
The browser calls `GET /api/carts/{id}` on load, every minute and on tab focus, drops
lines the server no longer holds and shows "Krepšelis atnaujintas". Each line shows
"Rezervuota iki HH:MM". A background refresh is ignored if a cart change ran meanwhile.
Rejected: Firestore TTL policy (deletes the doc but can't put stock back);
lazy expiry on the next request (stock stays blocked if nobody calls the API).
Note: Render's free tier sleeps when idle, so the sweep runs once the API wakes.

## 17. `.env.example` files restored
They were deleted in 684c9f0 but README/CLAUDE.md still pointed to them. Restored as
templates; real `backend/.env` and `frontend/.env.local` are gitignored.

## 18. Stay on Firestore; search stays in Go
Question raised: Firestore has no full-text/fuzzy search, so should we switch databases?
No. Search never used the database: the service caches the product list and does
trigram matching in Go (#2, #3), so it works the same on any store. At tens to low
thousands of products that is fast enough.
If the catalog grew to ~100k+ products, the move would be PostgreSQL with the `pg_trgm`
extension (trigram similarity + GIN index inside the database; free hosts: Supabase,
Neon), or a search engine (Typesense/Meilisearch) next to Firestore.
Rejected for now: migrating to Postgres (rewrite of the repository layer and stock
transactions, a new `pgx` dependency, new hosting, for no visible gain at this size);
a separate search service (one more thing to deploy and keep in sync).
