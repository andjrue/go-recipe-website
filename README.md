# go-recipe-website

A private, shared cookbook. Go backend (repository pattern over Postgres, pgx driver),
React frontend (later), deployed to AWS (later).

- **Architecture & data model:** [docs/notes/structure.md](docs/notes/structure.md)
- **Progress / dev log:** [docs/notes/devlog.md](docs/notes/devlog.md)

## Requirements

- Go 1.25+
- Docker (for local Postgres)
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
make migrate-up   # apply migrations
make run          # start the app
make db-down      # stop Postgres when you're done (data persists in a volume)
```

## Migrations (goose)

Migrations live in `internal/database/migrations/` and are the **single source of truth**
for the schema.

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
| `make tidy` | `go mod tidy` |
| `make fmt` | `go fmt ./...` |
