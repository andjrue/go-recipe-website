-- +goose Up
-- +goose StatementBegin

ALTER TABLE recipes DROP CONSTRAINT IF EXISTS fk_recipes_user;

ALTER TABLE recipes
    ALTER COLUMN recipe_id DROP DEFAULT,
    ALTER COLUMN recipe_id TYPE UUID USING recipe_id::uuid,
    ALTER COLUMN user_id type UUID USING user_id::uuid;

ALTER TABLE users
    ALTER COLUMN user_id DROP DEFAULT,
    ALTER COLUMN user_id type UUID USING user_id::uuid;

ALTER TABLE recipes
    ADD CONSTRAINT fk_recipes_user
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes
DROP TABLE IF EXISTS users
-- +goose StatementEnd
