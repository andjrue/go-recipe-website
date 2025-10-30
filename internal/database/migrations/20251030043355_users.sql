-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    hashed_password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    alias VARCHAR(25) NOT NULL UNIQUE CHECK (length(alias) > 3),
    date_joined TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_alias ON users(alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
