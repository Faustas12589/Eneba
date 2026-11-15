# API

All responses are JSON. Prices are integer cents. Errors: `{"error": "..."}`.

## Public
| Method | Path | Notes |
|---|---|---|
| GET | `/healthz` | liveness (Render health check) |
| GET | `/api/products` | query: `q`, `category`, `platform`, `region`, `genre` (comma-separated; `category` is one of `games`, `dlc`, `gaming-cards`, `gift-cards`, `software`), `minPrice`, `maxPrice` (cents), `sort` (`popular`/`price_asc`/`price_desc`/`discount`/`name`), `page`, `pageSize` (max 60). Returns `{items,total,page,pageSize,facets}` |
| GET | `/api/products/{id}` | `{product, offers[]}`, offers cheapest first |
| GET | `/api/search?q=` | top 6 suggestions for the header dropdown |
| GET | `/api/home/banners` | homepage block: banners |
| GET | `/api/home/best-sellers` | homepage block: top by sales |
| GET | `/api/home/gift-cards` | homepage block: best-selling `gift-cards` |
| GET | `/api/home/top-ups` | homepage block: best-selling `gaming-cards` (Steam, PSN, Roblox...) |
| GET | `/api/home/pre-orders` | homepage block: products with a future `releaseDate`, soonest first |
| GET | `/api/home/categories` | homepage block: `{items: [{category, count}]}` in menu order |

## Cart (anyone holding the cart id, DECISIONS #15, #19)
| Method | Path | Notes |
|---|---|---|
| GET | `/api/carts/{cartId}` | the cart order as stored; 410 once paid |
| PUT | `/api/carts/{cartId}/items/{offerId}` | body `{quantity}` (0 = remove). Reserves stock + keys. `{cart, offerId, stock}`; 409 not enough stock, 410 paid |

## Guest user (header `X-Guest-Id: <random id>`, while login is off: DECISIONS #20)
Missing or malformed id -> 401.

| Method | Path | Notes |
|---|---|---|
| POST | `/api/orders` | body `{cartId}`: simulated payment, then delivery. 201 delivered order with key codes. Safe to retry |
| GET | `/api/orders?page=&pageSize=` | `{items,total,page,pageSize}`, newest first (DECISIONS #22) |
| POST | `/api/profile/views` | body `{productId}`: adds it to the profile's recently viewed. 204; 404 unknown product |
| GET | `/api/home/recently-viewed` | homepage block: `{items}` up to 8, newest first |
| GET | `/api/home/recommended` | homepage block: `{items}` 8 products scored by genre/category/platform overlap with views + purchases (DECISIONS #25) |

## Signed-in (header `Authorization: Bearer <Firebase ID token>`)
With `STORE=memory` the only accepted token is `dev` (user `dev-user`), see DECISIONS #10.

| Method | Path | Notes |
|---|---|---|
| GET | `/api/me` | `{uid,email}` |
