-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipes
ADD CONSTRAINT fk_recipes_user
FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recipes
DROP CONSTRAINT IF EXISTS fk_recipes_user;
-- +goose StatementEnd
