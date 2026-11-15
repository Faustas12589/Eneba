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
| GET | `/api/home/deals` | homepage block: top by discount |

## Signed-in (header `Authorization: Bearer <Firebase ID token>`)
With `STORE=memory` the only accepted token is `dev` (user `dev-user`), see DECISIONS #10.

| Method | Path | Notes |
|---|---|---|
| GET | `/api/me` | `{uid,email}` |
| POST | `/api/orders` | TODO: create order from cart, simulate payment, return keys |
| GET | `/api/orders` | TODO: order history |
