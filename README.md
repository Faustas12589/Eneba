# Eneba-style Marketplace

Digital game-key marketplace (internship assignment). React + TypeScript frontend,
Go REST backend, Firebase (Firestore + Auth), deployed on Render.

**Live:** _add URL_ · **Repo:** _add URL_ · See `SPEC.md` (scope), `DECISIONS.md` (trade-offs), `AI_LOG.md` (agent history).

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
go mod tidy          # generates go.sum (required once, commit it)
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
1. Create a Firebase project; enable **Firestore** and **Authentication > Email/Password**.
2. Project settings > Service accounts > generate a private key (keep it OUT of git).
3. Project settings > Your apps > add a Web app; copy config into `frontend/.env.local` (see `.env.example`).
4. Seed and run the backend against Firestore:
   ```bash
   cd backend
   export FIREBASE_PROJECT_ID=your-project
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json
   go run ./cmd/seed
   STORE=firestore go run ./cmd/api
   ```

## Deploy (Render)
1. Push to GitHub, then Render > New > Blueprint > pick the repo (uses `render.yaml`).
2. Set the `sync: false` env vars: `FIREBASE_PROJECT_ID`, `FIREBASE_CREDENTIALS_JSON` (paste the whole JSON), `CORS_ORIGINS` (frontend URL), `VITE_*` (API URL + Firebase web config).
3. In Firebase Auth > Settings > Authorized domains, add the Render frontend domain.
4. Note: the free API service sleeps when idle; the first request can take ~30-60s.

## Status
Done (scaffold): catalog, filters/sort/search, homepage blocks, product page with offers, auth plumbing.
TODO: cart, checkout + key delivery, order history, Eneba styling pass. See `SPEC.md`.
