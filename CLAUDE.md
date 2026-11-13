# Project: Eneba-style marketplace (internship assignment)

Digital-key marketplace that looks and feels like eneba.com. Graded on: working
features, how I use AI, and whether I can explain every line. Read SPEC.md for
scope and DECISIONS.md for past trade-offs before changing anything.

## Always-loaded context
@SPEC.md
@DECISIONS.md
@docs/DESIGN.md

## Stack
- Frontend: React 19 + TypeScript + Vite, `react-router-dom`, feature folders (`frontend/src/features/*`)
- Backend: Go, `chi` router, layered: `handler -> service -> repository` (`backend/internal/*`)
- DB/Auth: Firebase Firestore + Firebase Auth (backend verifies ID tokens)
- Deploy: Render (Go web service via Docker, frontend as static site) — see `render.yaml`

## Commands
- First-time setup: `cd backend && go mod tidy` (writes go.sum, commit it) and `cd frontend && npm install`
- Backend dev: `cd backend && go run ./cmd/api` (STORE=memory by default, no Firebase needed)
- Backend tests: `cd backend && go test ./...`
- Seed Firestore: `cd backend && go run ./cmd/seed` (needs `FIREBASE_PROJECT_ID` + `GOOGLE_APPLICATION_CREDENTIALS`; idempotent)
- Frontend dev: `cd frontend && npm run dev` (proxies /api to :8080)
- Frontend check: `cd frontend && npm run build` (runs tsc then vite build)

## Rules
1. Layering: handlers only parse/serialize HTTP; business logic lives in `service`;
   only `repository` touches Firestore. Services depend on interfaces.
2. Money is integer in cents everywhere. Format only in the UI (`lib/format.ts`).
3. Homepage blocks: one backend endpoint per block, frontend renders what it receives.
   Nothing hardcoded in the frontend.
4. All colors/spacing/fonts come from `frontend/src/styles/tokens.css` CSS variables.
   Never hardcode a color in a component. Visual reference: `docs/DESIGN.md`.
5. Never commit secrets (service-account JSON, `.env`). Only `.env.example`.
6. Work in small vertical slices. Run tests/build before saying a slice is done. One commit per slice.
7. Keep code simple enough - easy to explain. Prefer boring over clever.
   After each slice, add a short entry to DECISIONS.md (what, why, alternatives rejected).
8. Ask before adding a new dependency.

## AI log (required deliverable)
The full chat history (prompts AND responses) must be submitted. At the end of each
session, run `/export` and append the result to `AI_LOG.md` unedited, with a date heading.
