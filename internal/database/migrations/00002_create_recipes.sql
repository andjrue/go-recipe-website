-- +goose Up
-- +goose StatementBegin
CREATE TABLE recipes (
    recipe_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL CHECK (btrim(name) <> ''),
    recipe_type    TEXT NOT NULL CHECK (recipe_type IN ('structured', 'image')),
    time_to_cook   TEXT NOT NULL DEFAULT '',
    description    TEXT NOT NULL DEFAULT '',
    user_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    date_posted    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_edited_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX recipes_active_date_posted_idx
    ON recipes (date_posted DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE ingredients (
    ingredient_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id     UUID NOT NULL REFERENCES recipes(recipe_id) ON DELETE CASCADE,
    name          TEXT NOT NULL CHECK (btrim(name) <> ''),
    quantity      TEXT NOT NULL DEFAULT '',
    position      INTEGER NOT NULL CHECK (position >= 0),
    UNIQUE (recipe_id, position)
);

CREATE TABLE recipe_steps (
    step_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id   UUID NOT NULL REFERENCES recipes(recipe_id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL CHECK (step_number > 0),
    instruction TEXT NOT NULL CHECK (btrim(instruction) <> ''),
    UNIQUE (recipe_id, step_number)
);

CREATE TABLE recipe_images (
    image_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id    UUID NOT NULL REFERENCES recipes(recipe_id) ON DELETE CASCADE,
    s3_key       TEXT NOT NULL UNIQUE CHECK (btrim(s3_key) <> ''),
    file_name    TEXT NOT NULL CHECK (btrim(file_name) <> ''),
    content_type TEXT NOT NULL CHECK (content_type IN ('image/jpeg')),
    file_size    BIGINT NOT NULL CHECK (file_size > 0),
    position     INTEGER NOT NULL CHECK (position >= 0),
    is_cover     BOOLEAN NOT NULL DEFAULT false,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (recipe_id, position)
);

CREATE UNIQUE INDEX recipe_images_one_cover_idx
    ON recipe_images (recipe_id)
    WHERE is_cover;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE recipe_images;
DROP TABLE recipe_steps;
DROP TABLE ingredients;
DROP TABLE recipes;
-- +goose StatementEnd
