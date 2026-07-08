# Dev Log

A running record of what's built, what's decided, and what's next — so picking this
back up (in a new session or solo) doesn't mean re-deriving everything. Newest entries
at the top.

For the *why* behind the architecture, see [structure.md](./structure.md). This file is
the "where are we" companion to that "what are we building" doc.

---

## Status at a glance

| Area | State |
|------|-------|
| Project decisions / data model | ✅ Settled (see structure.md) |
| Local Postgres (Docker) | ✅ Working |
| Migrations (goose) | ✅ Working — Users table live |
| DB connection (pgx pool) | ✅ Working |
| Users repository | ✅ Interface + full Postgres impl, verified |
| HTTP handlers | ✅ Router + users read handler |
| Recipes / Ingredients / Steps / Images | ⬜ Not started |
| Auth (Google OAuth) | ⬜ Not started |
| Frontend (React) | ⬜ Not started |
| AWS / Terraform | ⬜ Not started |

---

## 2026-07-05 — HTTP foundation: router + users handler

**Built & verified:**
- Added `internal/api` with a standard-library `http.ServeMux` router.
- Added `GET /api/health` and `GET /api/users/{id}`.
- Wired `cmd/main.go` into an HTTP server using `HTTP_ADDR` with a `:8080` default.
- Added API response helpers and handler tests with a fake `UserRepository`.
- Documented handler/router conventions in [http.md](./http.md).

---

## 2026-07-05 — Review cleanup: users foundation

**Adjusted after review:**
- Removed the early GitHub Actions workflow. We'll add CI back once the repo has a clearer
  test/build shape.
- Kept `.env.example` as committed setup documentation with non-secret local defaults.
- Tightened the users migration: `alias` is now non-null with an empty-string default,
  `provider` is constrained to `google`, and `role` is constrained to `user` / `admin`.
- `UserRepository` now maps Postgres/pgx details into shared application errors
  (`apperror.ErrNotFound`, `apperror.ErrConflict`) so handlers don't need to import pgx.
- `Create` now defaults an empty provider to `google` in Go before insert, with the DB
  default still present as a backstop.

---

## 2026-07-03 — Foundations: decisions + Users vertical slice

**Decisions locked in** (full detail in structure.md):
- Scope trimmed to a **private cookbook** for v1. Comments, ratings, and public/family
  multi-user are deferred to a "One Day" list.
- **Auth:** Google OAuth (OIDC) — we store no passwords. Users carry `provider` /
  `provider_user_id` and a `role` field (the seam for going multi-user later).
- **Recipes are either/or** via a `RecipeType` discriminator: `structured`
  (ingredients + steps) or `image` (uploaded photos).
- **Image uploads normalize** to one web format on the way in; multiple images per
  recipe, each with a `Position` and an `IsCover` flag.
- **Stack:** Go stdlib handlers + hand-written repository pattern (no ORM, no sqlc —
  we want mockable interfaces and full control). Postgres via **pgx/v5**. Migrations
  via **goose**. React frontend, separate app. Deploy target: ECS Fargate + RDS,
  all via Terraform.

**Built & verified end-to-end:**
- Swapped `lib/pq` → `pgx/v5`; connection is now a `pgxpool.Pool`
  (`internal/database/postgres.go`).
- First goose migration: `users` table with UUID PK, a unique `(provider,
  provider_user_id)` pair for OAuth lookups, and sensible defaults.
- `UserRepository` interface + `UserPostgres` implementation (`Create`, `GetByID`,
  `GetByProviderUserID`) — hand-written SQL, compile-time interface check.
- Local dev infra: `docker-compose.yml` (Postgres 17), `.env` / `.env.example`,
  `.gitignore`, and Makefile targets for db + migrations.
- Verified the whole loop: `make db-up` → `make migrate-up` → `make run` connects,
  and column defaults (`gen_random_uuid()`, `role='user'`, `provider='google'`) fire
  correctly on insert.

**Loose ends noticed (not yet addressed):**
- `TimeToCook` is modeled as a free-text string — flexible, but useless for
  sorting/filtering. Fine for v1.
- "Exactly one cover image per recipe" isn't enforced yet — plan is a Postgres partial
  unique index (`UNIQUE (recipe_id) WHERE is_cover`) when we build the Images table.

---

## Next up (pick one)

1. **Recipes vertical** — migrations + repositories for Recipe / Ingredients / Steps /
   RecipeImage, including the `RecipeType` discriminator and the cover-image constraint.
2. **Google OAuth** — the login flow that actually populates the users table.

## One Day (deferred on purpose)
Comments, ratings, family/multi-user, and LLM/OCR (photo → editable structured recipe).
Tracked in structure.md.
