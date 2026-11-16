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

## 11. Digital keys: seeded pool per offer (built differently, see #19)
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

## 19. The cart is an order; keys are reserved from the pool at cart time (supersedes #11, storage part of #15)
Two new collections. `keys`: `{offerId, code, status, orderId}`, status
`available -> reserved -> sold`. `orders`: status `cart -> paid -> delivered`. A cart is
simply an order in status `cart` (doc id = the browser's cartId), so one document goes
from basket to receipt, and a key's `orderId` points at it the whole way.
- Add to cart (#15 rule unchanged): the same transaction moves stock, updates the cart's
  line and flips that many keys `available -> reserved` (orderId = cart). Removing or
  expiring (#16) flips them back. So a key can only be taken from the unreserved pool,
  and `offers.stock` always equals the offer's available keys (seed: exactly `stock`
  keys per offer, generated in `seeds/keys.go`). The counter stays so the product page
  doesn't have to count keys. Tests check stock == available keys after every step.
- Checkout `POST /api/orders {cartId}` (signed in), two transactions:
  pay (cart -> paid: owner, server prices, total) then deliver (paid -> delivered: the
  reserved keys become sold, codes copied into the order as a receipt). Delivery can't
  run out because the keys were reserved earlier. Retrying finishes a paid order or
  returns the delivered one; another user's cart answers 404.
- A paid cart is closed: changing it returns 410, and the browser starts a new cartId.
- Rules (`pay`, `deliver`, `reserve`) are pure functions in `service`; repositories run them atomically.
Rejected: claiming keys only at checkout (#11's plan: reservation would be a counter
without real keys behind it); keeping a separate `reservations` collection next to orders
(two places describe the same cart); one transaction for pay + deliver (the `paid`
stage would never be visible, and a failed delivery would undo the payment).
Trade-offs: everyone adding the same offer queries the same first available keys, so
Firestore retries under heavy contention (fine at demo traffic). Seeded codes are
deterministic, i.e. fake. Re-seeding wipes carts and orders.

## 20. Login switched off for now: orders belong to a browser guest id
Checkout and order history no longer need sign-in. The browser makes a random id once
(`lib/guest.ts`, localStorage) and sends it as `X-Guest-Id`; `middleware.RequireGuest`
turns it into the user `guest:<id>`, so handlers, services and the `orders.userId` field
are unchanged. Same trust model as cart ids (#15): long, random, unguessable.
Why: real Firebase sign-in blocked testing on Firestore, and the dev token (#10) only
works in memory mode. Undo: use `RequireAuth(verifier)` for the orders group in
`router.go`, wrap `/checkout` and `/orders` in `<RequireAuth>`, restore the header login
button. The Firebase auth code is kept.
Trade-off: orders live in one browser; clearing storage or switching device loses the
history (the orders stay in Firestore). Rejected: one shared fake user (every visitor
would see everyone's keys).

## 21. Order data is written once; stages only change status fields (refines #19, supersedes #13)
When a line goes into the cart, the server writes it whole into the cart order: product
name, image, seller, price (snapshot from that moment), quantity, `expiresAt` and the
reserved keys (`{id, code}`); `total` is updated on every cart change. After that:
- pay (cart -> paid) only sets `status`, `userId`, `paidAt` and clears `expiresAt`;
- deliver (paid -> delivered) only sets `status`, `deliveredAt` and marks the keys sold.
Checkout never looks up products or prices, so it can't "repair" or change data
(test: `TestCheckoutOnlyChangesState`; checked on Firestore by diffing the document).
- `paidAt`, `deliveredAt`, `expiresAt` are nullable (`*time.Time`): null until that
  stage, instead of Go's zero date `0001-01-01`.
- Key codes sit on the order from cart time, but the API shows them only once the order
  is delivered (`Order.Public()`), so a cart can't leak free keys. Firestore is
  backend-only (#12), so storing them there is safe.
- Releasing units gives back the line's last keys by id, no query needed.
- The browser renders the cart from the server's order (`GET /api/carts/{id}`, and every
  `PUT` returns the stored cart). localStorage keeps only the cart id (supersedes #13's
  browser-side snapshot). A paid cart answers 410 and the browser starts a new one.
Trade-off: the price is fixed when the line is added, so a later price change does not
affect carts already holding it (like a reservation). That replaces #7's "re-read prices
at checkout"; still safe because only the server writes the price.
Rejected: filling names/prices at payment (data appears late and checkout does two jobs);
keeping a browser copy of the lines (two sources of truth that drift).

## 22. Order history is paged in the service, like the store
`GET /api/orders?page=&pageSize=` returns `{items, total, page, pageSize}` (default 10, max 50),
the same shape as `/api/products`. The service reads all of the user's orders, sorts them
newest first (ties broken by id so pages never overlap) and cuts the page with
`pageBounds`, which the catalog now shares. The frontend has one `Pager` component for both
pages; the page number lives in the URL (`/orders?page=2`), so back/forward works.
Why in Go: a shopper has tens of orders, and a Firestore `where userId ==` + `orderBy paidAt`
query needs a composite index to deploy and explain.
Rejected: Firestore `limit` + cursor paging (index + cursor tokens in the API; only worth it
for users with thousands of orders); `offset` in Firestore (still bills the skipped reads).

## 23. Every product has a category (product type), separate from genre
Eneba's menu and filters start from the product type: games, DLC, gaming cards
(Steam/PSN/Xbox top-ups), gift cards, software. `Product.category` holds one of the fixed ids in
`domain.Categories`; genre stays for games and DLC and is `""` for the rest (the seed used
to fake this with the genre "Gift Card"). The store filters by `?category=a,b` like the other
facets, the facet list keeps the menu order, and the header has a category row linking to
`/store?category=…`. The frontend only maps ids to Lithuanian labels (`lib/categories.ts`).
Seed grew from 16 to 24 products (2 DLC, 2 gaming cards, 2 gift cards, 2 software) with small
stock (+357 keys) so each category has entries; a seed test checks every category is known
and `priceFrom` equals the cheapest offer.
Rejected: a `categories` collection in Firestore (5 fixed values need no reads or admin UI);
several categories per product (Eneba shows one type per item; tags can come later);
reusing genre (a card has no genre, and a game needs both).
Firestore needs a re-seed (`go run ./cmd/seed`) to get the field; it wipes carts/orders (#19).

## 24. Eneba styling from their real CSS, all values in tokens.css
Colors, font and sizes come from eneba.com's own stylesheets (fetched with curl and counted by
frequency), checked against screenshots in `docs/screens/`; `docs/DESIGN.md` lists each value
and where Eneba uses it. `tokens.css` holds every value; components only use the variables
(no hex outside tokens.css). Metropolis (SIL OFL) is self-hosted from `public/fonts/` with
`@font-face`, only the 4 weights used (400/600/700/800).
Layout follows Eneba: two-row header (search + category band), narrow cards with platform strip,
green region, "Nuo" + struck list price, single-row scrolling rails on the homepage, cart as
"Mano krepšelis" + "Suvestinė" panels with a yellow CTA. Region codes get Lithuanian labels
(`regionLabel`). Cards without an image show the product name on a gradient.
Checked visually with headless Chrome screenshots at 1280px and 500px.
Rejected: copying Eneba's logo/images (brand assets, not ours: the logo is plain text);
a CSS framework (Tailwind etc.: new dependency, and tokens already do the job);
Google Fonts Montserrat (close, but the real font is free to self-host).
Follow-up: on phones the store filters should collapse behind a button like Eneba's.

## 25. Profiles per user id: recently viewed + recommendations
`profiles/{userId}` stores the user's last 20 viewed product ids (newest first). The user id is
whatever the request's user is: `guest:<id>` while login is off (#20), the Firebase uid once it is
back, so nothing needs rewriting then. Purchases are not copied into the profile; they are read
from the user's orders (one source of truth).
- The product page calls `POST /api/profile/views` (fire and forget). Reloading the product that
  is already first skips the write (Firestore quota).
- `GET /api/home/recently-viewed` and `GET /api/home/recommended` are their own endpoints like
  every homepage block (#6), behind `RequireGuest`. Recently viewed is hidden while empty.
- `recommend` is a pure function: each product not viewed or bought scores +3 per history item
  with the same genre, +2 same category, +1 same platform; ties and new users fall back to sales.
  `addView` (move to front, dedupe, cap) is pure too. Both are unit tested.
Rejected: login first (blocked Firestore testing before; Eneba shows these blocks to guests too);
storing views in localStorage (recommendations need purchases, which live on the server, and a
profile can follow the user once login returns); collaborative filtering ("people who bought X
also bought Y": needs far more orders than a demo has, and is harder to explain).
Trade-off: one Firestore write per product page view (~fine at demo traffic, 20k/day free).
When login returns, a guest's profile could be merged into the account; not built yet.

## 26. Homepage = Eneba's sections; pre-orders from a release date; smaller seed
Blocks, top to bottom: banners, recently viewed, recommended (#25), best sellers, popular gift
cards, popular top-ups, pre-orders, categories with counts. "Geriausi pasiūlymai" (deals) is gone
(endpoint too). Each block keeps its own endpoint (#6); the two "popular" rails reuse one service
method (`TopInCategory`) with a different category. Every block is a full-width band and CSS
alternates the colors (`.band:nth-child(even)`), so a hidden block (empty recently viewed) never
breaks the pattern.
- Pre-orders: `Product.releaseDate` ("YYYY-MM-DD", stored); `preOrder` is derived when the catalog
  loads (never stored, like the discount in #9), so GTA VI (2026-11-19) turns into a normal product
  on its own on release day. ISO dates compare as text, so no date parsing. The catalog's clock is
  injectable so the test doesn't break once the seeded dates pass. Pre-orders are sold like any
  key (delivered at once); a real shop would hold delivery until release.
- Seed: every offer's stock divided by 20 (5,863 -> 322 keys), 1-10 units per offer, so selling
  out, 409s and the 30-min expiry can be tried by hand; a Firestore seed is now a few hundred writes.
  Added a Roblox card (top-ups) and 3 pre-orders (release dates are demo data).
Rejected: a stored `isPreOrder` flag (goes stale on release day); one generic
`/api/home/block?type=` endpoint (breaks "one endpoint per block").

## 27. Second design pass from a side-by-side screenshot
Matched `docs/screens/desired.png` (Eneba) against `current.png`: icon header (cart with red
badge, avatar linking to orders), "☰ Kategorijos" row, full-bleed hero (first banner spans 3
grid columns, the others 1), category icon row with counts moved under the hero, 32px section
titles, 214px cards, thin themed scrollbar. Icons are hand-written inline SVG in
`components/Icons.tsx` using `currentColor`, so colors still come from tokens; one new token
(`--color-badge`).
Skipped on purpose: wishlist heart and language/currency picker (no such features, a dead icon
would only be decoration); Eneba's logo and banner art (their brand assets).
Rejected: an icon library such as lucide-react (new dependency for 9 icons).

## 28. Real login is back: guests do everything except pay (supersedes #20)
Firebase Auth with email/password and Google (`signInWithPopup`, part of the `firebase` package
already installed). Guests keep their browser id (#20) and can browse, fill a cart, see recently
viewed, recommendations and their guest orders. Only paying needs an account: pressing
"Apmokėti" as a guest opens a "Prisijungimas būtinas" popup, and payment continues right after.
- Routes: `middleware.Identify` uses the login token if one is sent, else `X-Guest-Id`
  (orders list, profile views, personal blocks). A bad token is a 401, never silently a guest.
  `RequireAuth` guards checkout, `GET /api/me` and `POST /api/me/merge-guest`.
- Merge on sign-in: the guest's recently viewed go in front of the account's (pure `mergeViews`,
  deduped, max 20) and the guest's orders get the account's user id (`MoveUserOrders`, one
  transaction). Safe to repeat. The browser runs it before switching `user`, so blocks that reload
  on sign-in already see the merged data.
- User info: the header avatar is the Google photo, or the plain user icon (no initials). Its menu
  shows name + email and links to `/account` (name editable, email, sign-in method, member since,
  order count, total spent). Name/photo/dates come from Firebase Auth in the browser; only order
  numbers come from the API, since orders live there.
Rejected: login required for the cart (more friction, and Eneba lets guests shop);
starting accounts empty (a guest who signs in mid-shop would lose their history);
initials avatars (the user asked for the picture only); Firebase UI library (new dependency).
Trade-off: whoever holds a guest id can merge that guest's data into their account, the same
trust model as reading those orders with the id (#20). Merging is not undone on sign-out.

## 29. Third design pass: login page, category drawer, icon row, section info pop-ups
From the new screenshots in `docs/screens/` (header logged in/off, log in page, categories side
bar, info pop up):
- Header: signed out shows "👤 Prisijungti | Registruotis" (register = `/login?mode=register`);
  signed in shows the avatar menu (#28). Category row links use the new `--color-text-nav`.
- `/login` is its own page outside the shop layout, like Eneba's: greeting on the left, panel on
  the right, white Google button with Google's G (their branding asks for it; colors are tokens).
  The same `SignInForm` is used in the checkout popup.
- "☰ Kategorijos" opens a left drawer (`CategoryDrawer`): "Visi pasiūlymai", the 5 categories,
  then the platforms from the store's facets (no hardcoded platform list).
- Icon row: name only (the count is gone from the UI; the endpoint still returns it), smaller and
  tighter. Our 5 categories repeat to fill the row (asked for the look), and a reusable
  `ScrollRow` adds ‹ › buttons that only show when there is more to scroll that way.
- Rails: "Rodyti visus" is an outlined button under the cards. Every section title has an ⓘ
  (`InfoTip`) whose dark pop-up says what the block shows and how it differs for guests vs
  signed-in users. Shown with CSS on hover/focus, no state.
Skipped on purpose: Facebook/Apple/Steam/Discord buttons and magic-link email (not built, a dead
button would mislead); wishlist heart and language picker (as in #27); drawer sub-menus behind
the › arrows (each row links straight to the store filter).
Rejected: a tooltip library (new dependency for one pop-up); fake categories to fill the icon row
(they would lead to empty store pages).

## 30. Store sidebar: price range slider instead of the "Kategorija" checkboxes
The category already comes from the header row, the side menu (#29) and the homepage icons, so
the sidebar's first section is now a price range: a two-handle slider plus "Nuo/Iki" boxes in
whole euros (URL and API stay in cents, rule 2). The slider's ends come from the API:
`facets.priceMin/priceMax` are the cheapest and dearest products matching every filter except
price, so inside "Žaidimai" it spans the games, and moving the handles never shrinks the slider
itself (test: `TestListPriceRange`). The store reloads when a handle is let go or a box is left
or Entered, not on every keystroke like the old inputs; an empty box or a handle at the end means
no limit on that side.
Rejected: a slider library (two stacked native `<input type="range">` do it with ~30 lines of
CSS); fixed bounds such as 0-100 € (useless for cheap categories); keeping category checkboxes
as well (the user asked for them to go; `?category=` in the URL still filters).

## 31. Card hover buttons; compact side menu
Like Eneba, hovering a product card grows its info panel and shows a yellow "Pridėti į krepšelį"
and an outlined "Pirkti dabar". The cover slides up by exactly the buttons' height, so the card
keeps its size: nothing jumps and the scrolling rails don't clip it (pure CSS). The buttons sit
next to the link, not inside it (a button inside a link is invalid HTML). The list API has no
offer ids, so a click loads the product and adds its cheapest offer in stock
(`CartContext.addProduct`; the server still re-checks stock when reserving). "Pirkti dabar"
skips the popup and opens checkout. Touch screens (no hover) keep the plain card.
Side menu: smaller text, platforms in two columns under "Platformos", so it fits without scrolling.
Rejected: offer ids in every product list response (bigger payload for a hover action);
a card that grows downward (clipped by the rails' horizontal scroll).


## 32. Order history is signed-in only (refines #28)
The assignment says "checkout and order history are only for signed-in users". Since #28 guests
can't pay, so they never get new orders, but `GET /api/orders` still accepted a guest id. It now
sits in the `RequireAuth` group next to checkout, and `/orders` is wrapped in `<RequireAuth>` (a guest
is sent to `/login` and comes back after). Checkout keeps its #28 shape: guests see the summary,
"Apmokėti" opens the sign-in popup, which matches the assignment's flow diagram (Cart -> Checkout ->
sign-in required -> Payment). Old guest orders are not lost: they merge into the account on sign-in.
`router_test.go` checks each protected route: no token, guest id and bad token get 401, signed-in 200,
while the personal homepage blocks still accept a guest id.
Rejected: keeping guest order history and justifying it (the spec names it explicitly, and with
paying behind login it would only ever show an empty list).

## 33. Generated product covers instead of images (refines #24)
The seed has no cover art, and the real art belongs to Eneba or the publishers (#24 already
rejected copying it). The plain "name on one gradient" made every card look the same, so
`components/Cover.tsx` draws a cover when `imageUrl` is empty: one of 6 gradients from `tokens.css`,
a faint diagonal stripe, the category icon as a watermark, the category label and the title.
- The gradient is picked by `coverVariant(productId)` (`lib/cover.ts`): a plain string hash mod 6.
  The same product gets the same color on its card, product page, cart line and order (cart and
  order lines only know id + name, so they show the color and the first letter).
- Pure CSS (layered `background`s + custom properties), no images to host; a real `imageUrl`
  still wins, so adding art later is a data change only.
Rejected: Steam's CDN art (looks closest to Eneba, but not ours to use, and gift cards/top-ups have
none); random colors per render (a product would change color on every visit); an image service
such as placeholder.com (outside dependency at runtime for something CSS can do).
