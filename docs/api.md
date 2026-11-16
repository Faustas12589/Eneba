# API

All responses are JSON. Prices are integer cents. Errors: `{"error": "..."}`.

## Public
| Method | Path | Notes |
|---|---|---|
| GET | `/healthz` | liveness (Render health check) |
| GET | `/api/products` | query: `q`, `category`, `platform`, `region`, `genre` (comma-separated; `category` is one of `games`, `dlc`, `gaming-cards`, `gift-cards`, `software`), `minPrice`, `maxPrice` (cents), `sort` (`popular`/`price_asc`/`price_desc`/`discount`/`name`), `page`, `pageSize` (max 60). Returns `{items,total,page,pageSize,facets}` |
| GET | `/api/products/{id}` | `{product, offers[]}`, offers cheapest first |
| GET | `/api/search?q=` | top 6 suggestions for the header dropdown |
| GET | `/api/home` | homepage layout, top to bottom: `{blocks: [{id, type, personal?}]}`; `type` is `banners`, `categories` or `products`; each block loads `/api/home/{id}` (DECISIONS #35) |
| GET | `/api/home/banners` | homepage block: banners |
| GET | `/api/home/best-sellers` | homepage block: top by sales |
| GET | `/api/home/gift-cards` | homepage block: best-selling `gift-cards` |
| GET | `/api/home/top-ups` | homepage block: best-selling `gaming-cards` (Steam, PSN, Roblox...) |
| GET | `/api/home/pre-orders` | homepage block: products with a future `releaseDate`, soonest first |
| GET | `/api/home/categories` | homepage block: items `[{category, count}]` in menu order |

Every `/api/home/{id}` block answers `{id, title?, info?, moreUrl?, hideWhenEmpty?, items}`: the block's
text (title, the ⓘ text, the "Rodyti visus" link) comes with its items, so the frontend hardcodes none of it.

## Cart (anyone holding the cart id, DECISIONS #15, #19)
| Method | Path | Notes |
|---|---|---|
| GET | `/api/carts/{cartId}` | the cart order as stored; 410 once paid |
| PUT | `/api/carts/{cartId}/items/{offerId}` | body `{quantity}` (0 = remove). Reserves stock + keys. `{cart, offerId, stock}`; 409 not enough stock, 410 paid |

## Guest or signed-in (login token if sent, else header `X-Guest-Id: <random id>`; DECISIONS #28)
A bad login token is 401 (never silently a guest); a missing or malformed guest id is 401.

| Method | Path | Notes |
|---|---|---|
| GET | `/api/favorites` | the user's favorites (wishlist) `{items}`, newest first (DECISIONS #37) |
| PUT | `/api/favorites/{productId}` | add a favorite. 204; safe to repeat; 404 unknown product; 409 list full (100) |
| DELETE | `/api/favorites/{productId}` | remove a favorite. 204; safe to repeat |
| POST | `/api/profile/views` | body `{productId}`: adds it to the profile's recently viewed. 204; 404 unknown product |
| GET | `/api/home/recently-viewed` | homepage block: `{items}` up to 8, newest first |
| GET | `/api/home/recommended` | homepage block: `{items}` 8 products scored by genre/category/platform overlap with views + purchases (DECISIONS #25) |

## Signed-in only (header `Authorization: Bearer <Firebase ID token>`)
With `STORE=memory` the only accepted token is `dev` (user `dev-user`), see DECISIONS #10.
Anything else, including a guest id, is 401 (`router_test.go`).

| Method | Path | Notes |
|---|---|---|
| POST | `/api/orders` | body `{cartId}`: simulated payment, then delivery. 201 delivered order with key codes. Safe to retry |
| GET | `/api/orders?page=&pageSize=` | `{items,total,page,pageSize}`, newest first (DECISIONS #22, #32) |
| GET | `/api/me` | `{uid,email,orderCount,totalSpent}` (cents) |
| POST | `/api/me/merge-guest` | body `{guestId}`: moves that guest's views, favorites and orders to the account. Safe to repeat |
