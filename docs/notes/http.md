# HTTP Structure

The backend uses Go's standard library HTTP stack. Keep this layer small: it turns
requests into repository calls, maps application errors to HTTP status codes, and writes
JSON responses.

## Package Layout

- HTTP code lives in `internal/api`.
- Router construction happens in one place with `http.NewServeMux`.
- Resource handlers depend on repository interfaces, not concrete Postgres types.
- Shared application errors live in `internal/apperror`.

## Routing

Use method-aware standard-library patterns:

```go
mux.HandleFunc("GET /api/users/{id}", usersHandler.GetByID)
```

Handlers should read path parameters with `r.PathValue("name")`. Prefer resource routes
under `/api`. Health checks live at `/api/health`.

All resource routes are private unless explicitly documented otherwise. Protect them with
the auth handler's session middleware. Health checks and the Google login/callback endpoints
are the current public exceptions.

## Responses

Success responses return the plain resource JSON, not an envelope. Error responses use a
small consistent shape:

```json
{ "error": "not_found" }
```

Handlers should translate known `apperror` values into HTTP status codes. Unexpected
errors should return `500` with `internal_error` and avoid leaking implementation details
to the client. Log the underlying unexpected error on the server.

## DTOs

Keep request and response structs in `internal/api`. It is fine for early handlers to map
directly from repository models, but the HTTP boundary should own JSON field names and
should not expose database-specific types.

Do not expose identity-provider subject IDs. `/api/me` may return the signed-in user's email
and role; responses describing another user should use only the fields needed for display.
