# go-recipe-website

A private, shared cookbook. Go backend (repository pattern over Postgres, pgx driver),
React frontend, deployed to AWS (later).

- **Architecture & data model:** [docs/notes/structure.md](docs/notes/structure.md)
- **Progress / dev log:** [docs/notes/devlog.md](docs/notes/devlog.md)

## Requirements

- Go 1.25+
- Docker (for local API + Postgres)
- Node.js 22+ and npm
- [goose](https://github.com/pressly/goose) CLI for migrations:
  `go install github.com/pressly/goose/v3/cmd/goose@latest`

## Setup

```bash
cp .env.example .env   # then edit if needed; .env is gitignored
```

The `.env` holds DB connection config. Everything (app, migrations, docker) reads from it.

## Running locally

```bash
make db-up        # start Postgres 17 in Docker (detached)
make api-up       # build/start the API; pending migrations run automatically
make frontend-install
make frontend-dev # start the UI at http://localhost:5173
make db-down      # stop containers when you're done (data persists in a volume)
```

The API is available at `http://localhost:8080` by default:

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/ready  # also verifies the database connection
```

Uploaded recipe photos are normalized to JPEG and stored under `IMAGE_STORAGE_DIR`
(`data/images` for a host-run API and `/data/images` in Compose). Compose keeps them in the
named `recipe_images` volume; the database stores only their metadata.

`make run` still starts the API directly on the host for quick debugging. In Docker Compose,
the API service connects to Postgres through the internal service name `postgres`; host-run
migrations keep using the `.env` values.

If something else already owns `localhost:5432`, set `DB_PORT=5433` in `.env` before
starting Compose and running migrations. The API container will still connect to Postgres on
the internal Docker port.

The frontend development server proxies `/api` requests to `http://localhost:8080`, so the
browser does not need separate CORS configuration. For the browser login flow, set:

```bash
FRONTEND_URL=http://localhost:5173/recipes
```

## Google auth

Auth uses Google OIDC for identity and a signed HTTP-only cookie for the app session. There
is no sessions table; any backend container can verify the cookie with `SESSION_SECRET`.

Required `.env` values:

```bash
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
ALLOWED_EMAILS=you@example.com,partner@example.com
SESSION_SECRET=
SESSION_COOKIE_NAME=recipe_session
SESSION_COOKIE_SECURE=false
SESSION_TTL_HOURS=720
FRONTEND_URL=/api/me
```

`ALLOWED_EMAILS` is required and authentication stays disabled when it is empty. Use a
random `SESSION_SECRET` of at least 32 characters. In production, also set
`SESSION_COOKIE_SECURE=true` and serve the API over HTTPS.

Register `GOOGLE_REDIRECT_URL` as an authorized redirect URI in the Google OAuth client.
For local API-only testing, keep `FRONTEND_URL=/api/me` so a successful login redirects
straight to the current-user endpoint.

Test the flow in a browser:

```text
http://localhost:8080/api/auth/google/login
```

After Google redirects back, `/api/me` should return the signed-in user JSON. Logout with:

```bash
curl -i -X POST http://localhost:8080/api/auth/logout
```

## Migrations (goose)

Migrations live in `internal/database/migrations/` and are the **single source of truth**
for the schema. The API embeds and applies pending migrations during startup, so deployments
cannot serve against an outdated schema. The manual commands remain useful during development.

| Command | What it does |
|---------|--------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-status` | Show which migrations have run |
| `make migrate-create name=create_recipes` | Scaffold a new timestamped SQL migration |
| `make migrate-down` | ⚠️ **Roll back the most recent migration** |

> **⚠️ `migrate-down` is destructive.** It runs the `-- +goose Down` block of the latest
> migration, which typically `DROP`s tables and **deletes the data in them**. There is no
> "are you sure?" prompt. Only run it when you intend to reverse a schema change, and never
> against a database whose data you care about. When in doubt, run `make migrate-status`
> first to see exactly what's applied.

## Other Make targets

| Command | What it does |
|---------|--------------|
| `make build` | Compile to `bin/recipe-api` |
| `make test` | Run Go tests |
| `make tidy` | `go mod tidy` |
| `make fmt` | `go fmt ./...` |
| `make frontend-dev` | Start the React development server |
| `make frontend-test` | Run the Vitest suite |
| `make frontend-build` | Type-check and build the frontend |
| `make db-backup` | Write a timestamped Postgres dump under ignored `backups/` |
| `make images-backup` | Archive uploaded images under ignored `backups/` |
| `make backup` | Back up both Postgres and uploaded images |
| `make compose-up` | Build/start the API and Postgres containers |
| `make api-up` | Build/start only the API container and its dependencies |
| `make api-logs` | Follow API container logs |

Keep database dumps somewhere outside the project machine as well. Local dumps protect
against an accidental edit or migration; production RDS should also have automated snapshots
enabled when the AWS infrastructure is created.
