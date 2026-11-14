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
| Auth | Scaffolded | Firebase Auth email/password; backend verifies tokens. Dev mode without Firebase (DECISIONS #10). TODO: Google sign-in? (optional) |
| Cart | Done | Client-side cart, but stock is reserved on the server when an item is added (DECISIONS #15). Popup, header dropdown, live stock errors. |
| Checkout + key delivery | TODO | Simulated payment (DECISIONS #8). Keys come from a seeded pool per offer; stock = unsold keys. Transaction: claim key(s), create order (DECISIONS #11). Needs `Order`/`Key` types, `seeds/keys.json`, Firestore + memory orders repo. |
| Order history | TODO | Signed-in only. |
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
1. **Run on real Firestore.** Fill `backend/.env` + `frontend/.env.local`, `STORE=firestore`,
   `go run ./cmd/seed`. Check in the console: adding to cart lowers `offers.stock` and creates
   a `reservations` doc; it comes back after 30 min. (Transactions are only tested in memory so far.)
2. **Checkout + key delivery** (biggest slice). `POST /api/orders` (signed in): turns the
   cart's reservations into an order in one transaction, claims that many keys per offer from
   `seeds/keys.json`, deletes the reservations; the frontend clears the cart and shows the keys.
   Simulated payment (#8). First decide how #11 ("stock = unsold keys") fits #15 (stock is a
   counter reserved at cart time), e.g. seed exactly `stock` keys per offer so the counter and
   the key pool always agree. Needs `Order`/`Key` types, memory + Firestore repo, tests.
3. **Order history.** `GET /api/orders` (signed in) + a page listing past orders with their keys.
4. **Eneba styling.** Fill `docs/DESIGN.md` from eneba.com (colors, fonts, spacing, components),
   copy the values into `tokens.css`, polish header, cards, product page, cart dropdown, popup.
5. **Deploy to Render** with `render.yaml`, set env vars, test the live site end to end
   (free tier sleeps; the reservation sweep runs once it wakes).
6. **Tests + polish.** Handler tests for the cart endpoints; store/search edge cases; frontend
   unit tests for `cart.ts` (needs Vitest: ask first, rule 8); update README to match the code.
7. **Optional, if time:** 1-2 more homepage blocks (e.g. "Naujienos", a per-platform rail),
   Google sign-in.
8. **Every session:** `/export` the chat and append it to `AI_LOG.md` (required deliverable).
