# Scope & justification

> Edit this first. The assignment says scope is mine to decide and I must justify it.

## Principle
A few features done extensively beat many minimal ones. Every feature below has a
status; the "why" column is what I will say when asked.

| Area | Status | Depth / why |
|---|---|---|
| Store (filters, sort, search, pagination) | Scaffolded | Core of a marketplace. Filters + sort in Go (see DECISIONS #2). Aim: polish + tests. |
| Homepage blocks | Scaffolded | 3 blocks (banners, best sellers, deals), each own endpoint. Add 1-2 more if time. |
| Search | Scaffolded | Typo/partial tolerant (trigram), header dropdown + results via store page. |
| Product page + seller offers | Scaffolded | Offer comparison table with add-to-cart per offer. |
| Auth | Switched off | Firebase Auth email/password code is kept but not used: orders belong to a browser guest id for now (DECISIONS #20). Dev mode without Firebase (#10). TODO: turn back on; Google sign-in? (optional) |
| Cart | Done | The cart is a server-side order (status `cart`); the UI renders that document and the browser keeps only the cart id. Stock + keys reserved when an item is added (DECISIONS #15, #19, #21). Popup, header dropdown, live stock errors. |
| Checkout + key delivery | Done | Cart is the first stage of an order: cart -> paid -> delivered. Keys are reserved from the available pool when added to the cart, sold on delivery. Line data (name, price, keys) is written once at add time; checkout only changes status fields. Simulated payment (DECISIONS #8, #19, #21). |
| Order history | Done | Per browser guest id while login is off (DECISIONS #20); `GET /api/orders?page=`, paged 10 per page (#22), lists orders with copyable key codes. |
| Eneba styling | TODO | Tokens-driven (tokens.css is still placeholder). Extract real values into docs/DESIGN.md first. |

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
   Not yet seen live: the 30-min expiry sweep on Firestore. Cart changes take ~1.5-2s
   (transaction round trips); consider an optimistic UI if it feels slow.
   Login is switched off for now (DECISIONS #20); bring it back before or after deploy.
2. ~~Checkout + key delivery~~ Done (DECISIONS #19).
3. ~~Order history~~ Done.
4. **Eneba styling.** Fill `docs/DESIGN.md` from eneba.com (colors, fonts, spacing, components),
   copy the values into `tokens.css`, polish header, cards, product page, cart dropdown, popup.
5. **Deploy to Render** with `render.yaml`, set env vars, test the live site end to end
   (free tier sleeps; the reservation sweep runs once it wakes).
6. **Tests + polish.** Handler tests for the cart endpoints; store/search edge cases; frontend
   unit tests for `cart.ts` (needs Vitest: ask first, rule 8); update README to match the code.
7. **Optional, if time:** 1-2 more homepage blocks (e.g. "Naujienos", a per-platform rail),
   Google sign-in.
8. **Every session:** `/export` the chat and append it to `AI_LOG.md` (required deliverable).
