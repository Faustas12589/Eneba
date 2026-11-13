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
| Product page + seller offers | Scaffolded | Offer comparison table. TODO: add-to-cart per offer. |
| Auth | Scaffolded | Firebase Auth email/password; backend verifies tokens. Dev mode without Firebase (DECISIONS #10). TODO: Google sign-in? (optional) |
| Cart | TODO | Client-side cart of offers (localStorage, DECISIONS #7). |
| Checkout + key delivery | TODO | Simulated payment (DECISIONS #8). Keys come from a seeded pool per offer; stock = unsold keys. Transaction: claim key(s), create order (DECISIONS #11). Needs `Order`/`Key` types, `seeds/keys.json`, Firestore + memory orders repo. |
| Order history | TODO | Signed-in only. |
| Eneba styling | TODO | Tokens-driven (tokens.css is still placeholder). Extract real values into docs/DESIGN.md first. |

## Explicitly out of scope (and why)
- Real payments: assignment allows simulation; no value in wiring a provider.
- Seller dashboard / admin: not requested; seed data stands in for sellers.
- i18n, reviews, wishlists, email: not requested.

## Assumptions
- Currency: EUR only. Prices stored in cents.
- Catalog size: tens to low thousands of products, so in-memory filtering/search is fine.
