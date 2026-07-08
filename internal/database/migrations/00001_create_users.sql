-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    user_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email            TEXT NOT NULL UNIQUE,
    provider         TEXT NOT NULL DEFAULT 'google' CHECK (provider IN ('google')),
    provider_user_id TEXT NOT NULL,
    alias            TEXT NOT NULL DEFAULT '',
    role             TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    date_joined      TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- one account per identity from a given provider
    UNIQUE (provider, provider_user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
