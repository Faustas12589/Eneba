# Scope & justification

> Edit this first. The assignment says scope is mine to decide and I must justify it.

## Principle
A few features done extensively beat many minimal ones. Every feature below has a
status; the "why" column is what I will say when asked.

| Area | Status | Depth / why |
|---|---|---|
| Store (filters, sort, search, pagination) | Scaffolded | Core of a marketplace. Filters + sort in Go (see DECISIONS #2). Aim: polish + tests. |
| Homepage blocks | Done | 8 blocks, each its own endpoint, alternating bands: banners, recently viewed, recommended (#25), best sellers, gift cards, top-ups, pre-orders, category counts (#26). |
| Search | Scaffolded | Typo/partial tolerant (trigram), header dropdown + results via store page. |
| Product page + seller offers | Scaffolded | Offer comparison table with add-to-cart per offer. |
| Auth | Done | Firebase email/password + Google. Guests can do everything but pay; paying opens a sign-in popup, and guest views/orders merge into the account on sign-in (DECISIONS #28). Avatar menu + `/account` page. |
| Cart | Done | The cart is a server-side order (status `cart`); the UI renders that document and the browser keeps only the cart id. Stock + keys reserved when an item is added (DECISIONS #15, #19, #21). Popup, header dropdown, live stock errors. |
| Checkout + key delivery | Done | Cart is the first stage of an order: cart -> paid -> delivered. Keys are reserved from the available pool when added to the cart, sold on delivery. Line data (name, price, keys) is written once at add time; checkout only changes status fields. Simulated payment (DECISIONS #8, #19, #21). |
| Order history | Done | Signed-in only, as the assignment asks (DECISIONS #32); `GET /api/orders?page=`, paged 10 per page (#22), lists orders with copyable key codes. Guest orders from before login made it back merge into the account on sign-in (#28). |
| Eneba styling | Done | Values from eneba.com CSS in docs/DESIGN.md -> tokens.css; header, cards, store, product, cart restyled (DECISIONS #24, #27, #29, #31). Mobile filter drawer still TODO. Cards without art get a generated cover (#33). |

## Explicitly out of scope (and why)
- Real payments: assignment allows simulation; no value in wiring a provider.
- Seller dashboard / admin: not requested; seed data stands in for sellers.
- i18n, reviews, wishlists, email: not requested.

## Assumptions
- Currency: EUR only. Prices stored in cents.
- Catalog size: tens to low thousands of products, so in-memory filtering/search is fine (DECISIONS #18).
- UI language: Lithuanian only (matches eneba.lt). Seed data (names, descriptions) stays English.

## Next tasks (in order)
1. ~~Run on real Firestore~~ Done 2026-11-15: seeded 5559 keys; cart -> checkout -> history
   checked over HTTP, and stock == available keys for all 48 offers afterwards.
   Cart changes take ~1.5-2s (transaction round trips); consider an optimistic UI if it feels slow.
2. ~~Checkout + key delivery~~ Done (DECISIONS #19).
3. ~~Order history~~ Done; signed-in only since DECISIONS #32.
4. ~~**Eneba styling.**~~ Done (DECISIONS #24, #27, #29). Left: collapse store filters on phones. Covers are generated (#33).
5. ~~**Deploy to Render**~~ Done 2026-11-16: https://eneba-web.onrender.com (API `eneba-api-10wc`).
   Smoke-tested over HTTP: health, products, typo search, SPA routes, checkout 401 for guests.
   Not yet seen live: the 30-min expiry sweep (free tier sleeps; it runs once the API wakes).
6. ~~Login back~~ Done (DECISIONS #28), incl. Google sign-in.
7. **Tests + polish.** Handler tests for the cart endpoints (409/410); store/search edge cases;
   frontend unit tests for `cart.ts` (needs Vitest: ask first, rule 8). Router auth test done (#32).
8. **Optional, if time:** 1-2 more homepage blocks (e.g. "Naujienos", a per-platform rail).
9. **Every session:** `/export` the chat and append it to `AI_CHAT_LOG.md` (required deliverable).
