# Eneba-style Marketplace

Digital game-key marketplace (internship assignment). React + TypeScript frontend,
Go REST backend, Firebase (Firestore + Auth), deployed on Render.

**Live:** https://eneba-web.onrender.com (API: https://eneba-api-10wc.onrender.com/healthz) ·
**Repo:** https://github.com/Faustas12589/Eneba

See `SPEC.md` (scope and why), `DECISIONS.md` (trade-offs, numbered), `docs/api.md` (endpoints),
`docs/DESIGN.md` (Eneba styling reference) and `AI_CHAT_LOG.md` (full agent chat history).

> The free Render API sleeps when idle: the first page load after a while can take ~30-60 s
> while it wakes up. After that it is fast.

## Architecture
```
React (Vite)  --REST-->  Go API (chi)  -->  Firestore
   |                       handler -> service -> repository
   +-- Firebase Auth  --ID token-->  middleware verifies token
```
Backend layers: `handler` (HTTP) → `service` (logic, in-memory filter/search) → `repository` (Firestore or embedded seed data).
Frontend: feature folders under `frontend/src/features`, styling via CSS variables in `styles/tokens.css`.

## First-time setup

### 1. Backend (works without Firebase)
```bash
cd backend
go mod download      # go.sum is committed
go test ./...
go run ./cmd/api     # :8080, STORE=memory serves embedded demo data
```

### 2. Frontend
```bash
cd frontend
npm install
npm run dev          # http://localhost:5173, /api proxied to :8080
```
Without Firebase env vars the app runs in **dev auth mode**: any email/password signs
you in as a local dev user (only accepted by a `STORE=memory` backend). See `DECISIONS.md` #10.

### 3. Firebase (when ready)
1. Create a Firebase project; enable **Firestore** and **Authentication > Email/Password** and **Google**.
2. Project settings > Service accounts > generate a private key (keep it OUT of git).
3. Project settings > Your apps > add a Web app; copy config into `frontend/.env.local` (see `.env.example`).
4. Firestore > Rules: paste `firestore.rules` (deny all client access; only the Go API reads/writes).
5. Copy `backend/.env.example` to `backend/.env`, set `STORE=firestore`, `FIREBASE_PROJECT_ID`
   and `GOOGLE_APPLICATION_CREDENTIALS` (absolute path to the key, kept outside the repo).
6. Seed and run:
   ```bash
   cd backend
   go run ./cmd/seed    # idempotent full reset: products, offers, key pool; deletes carts/orders
   go run ./cmd/api     # log should say store=firestore
   ```

## Deploy (Render)
1. Push to GitHub, then Render > New > Blueprint > pick the repo (uses `render.yaml`).
2. Set the `sync: false` env vars: `FIREBASE_PROJECT_ID`, `FIREBASE_CREDENTIALS_JSON` (paste the whole JSON), `CORS_ORIGINS` (frontend URL), `VITE_*` (API URL + Firebase web config).
3. In Firebase Auth > Settings > Authorized domains, add the Render frontend domain.
4. Note: the free API service sleeps when idle; the first request can take ~30-60s.

## Status
Deployed and working end to end (store, search, product, cart, checkout, keys, order history).

| Part | What it does |
|---|---|
| Store | Filters (platform, region, genre, price slider), sort, typo-tolerant search, pagination; state in the URL. |
| Homepage | 8 blocks, each from its own endpoint (banners, recently viewed, recommended, best sellers, gift cards, top-ups, pre-orders, categories). |
| Search | Header dropdown from anywhere; trigram matching copes with typos and partial names (`witchr`, `eldne rign`). |
| Product page | Details + offers from several sellers, cheapest first, add to cart per offer. |
| Cart | Server-side: adding an item reserves real keys from the pool for 30 min; no overselling (transactions). |
| Checkout | Simulated payment, then the reserved keys are delivered and shown with copy buttons. |
| Auth | Firebase email/password + Google. Guests can browse and fill a cart; **paying and order history need an account** (sign-in popup at "Apmokėti"); guest data merges into the account on sign-in. |
| Styling | Colors, font (Metropolis), sizes from eneba.com's CSS in `tokens.css`; components use only those variables. |

Known gaps (see `SPEC.md`): no real cover images (cards show the name on a gradient), store filters do
not collapse on phones yet.

Firestore collections: `products`, `offers` (stock = available keys), `banners`,
`keys` (`offerId, code, status: available|reserved|sold, orderId`) and
`orders` (`status: cart -> paid -> delivered`; a cart is an order with status `cart`).
