-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS recipes (
    recipe_id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    time_to_cook TEXT,
    description TEXT,
    date_posted TIMESTAMP NOT NULL DEFAULT NOW(),
    last_edited_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_recipes_user_id ON recipes(user_id);
CREATE INDEX IF NOT EXISTS idx_recipes_recipe_id ON recipes(recipe_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;
-- +goose StatementEnd
