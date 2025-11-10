-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipes DROP COLUMN time_to_cook;
ALTER TABLE recipes ADD COLUMN time_to_cook INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;
-- +goose StatementEnd
