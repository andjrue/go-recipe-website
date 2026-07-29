package repository

import (
	"context"
	"fmt"

	"recipe-website/internal/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeRepository interface {
	Create(ctx context.Context, recipe *Recipe) (*Recipe, error)
	List(ctx context.Context) ([]*Recipe, error)
	GetByID(ctx context.Context, id string) (*Recipe, error)
	Update(ctx context.Context, recipe *Recipe) (*Recipe, error)
	Delete(ctx context.Context, id string) error
	AddImage(ctx context.Context, recipeID string, image *RecipeImage) (*RecipeImage, error)
	GetImage(ctx context.Context, imageID string) (*RecipeImage, error)
	DeleteImage(ctx context.Context, recipeID, imageID string) (*RecipeImage, error)
	SetCoverImage(ctx context.Context, recipeID, imageID string) error
}

type RecipePostgres struct {
	db *pgxpool.Pool
}

var _ RecipeRepository = (*RecipePostgres)(nil)

func NewRecipePostgres(db *pgxpool.Pool) *RecipePostgres {
	return &RecipePostgres{db: db}
}

func (r *RecipePostgres) Create(ctx context.Context, recipe *Recipe) (*Recipe, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning recipe create: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck -- rollback after commit is harmless

	const query = `
		INSERT INTO recipes (name, recipe_type, time_to_cook, description, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING recipe_id, date_posted, last_edited_at`

	err = tx.QueryRow(ctx, query,
		recipe.Name,
		recipe.RecipeType,
		recipe.TimeToCook,
		recipe.Description,
		recipe.UserID,
	).Scan(&recipe.ID, &recipe.DatePosted, &recipe.LastEditedAt)
	if err != nil {
		return nil, translatePostgresError(err)
	}

	if err := replaceStructuredRecipeRows(ctx, tx, recipe); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing recipe create: %w", err)
	}
	return recipe, nil
}

func (r *RecipePostgres) List(ctx context.Context) ([]*Recipe, error) {
	const query = `
		SELECT recipe_id, name, recipe_type, time_to_cook, description,
		       user_id, date_posted, last_edited_at,
		       COALESCE((SELECT image_id::text FROM recipe_images ri
		                 WHERE ri.recipe_id = recipes.recipe_id AND ri.is_cover LIMIT 1), '')
		FROM recipes
		WHERE deleted_at IS NULL
		ORDER BY date_posted DESC, recipe_id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing recipes: %w", err)
	}
	defer rows.Close()

	recipes := make([]*Recipe, 0)
	for rows.Next() {
		recipe, err := scanRecipe(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning recipe list: %w", err)
		}
		recipes = append(recipes, recipe)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recipe list: %w", err)
	}

	return recipes, nil
}

func (r *RecipePostgres) GetByID(ctx context.Context, id string) (*Recipe, error) {
	const query = `
		SELECT recipe_id, name, recipe_type, time_to_cook, description,
		       user_id, date_posted, last_edited_at,
		       COALESCE((SELECT image_id::text FROM recipe_images ri
		                 WHERE ri.recipe_id = recipes.recipe_id AND ri.is_cover LIMIT 1), '')
		FROM recipes
		WHERE recipe_id = $1 AND deleted_at IS NULL`

	recipe, err := scanRecipe(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, translatePostgresError(err)
	}

	if err := r.loadRecipeChildren(ctx, recipe); err != nil {
		return nil, err
	}
	return recipe, nil
}

func (r *RecipePostgres) Update(ctx context.Context, recipe *Recipe) (*Recipe, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning recipe update: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck -- rollback after commit is harmless

	const query = `
		UPDATE recipes
		SET name = $2,
		    recipe_type = $3,
		    time_to_cook = $4,
		    description = $5,
		    last_edited_at = now()
		WHERE recipe_id = $1 AND deleted_at IS NULL
		RETURNING user_id, date_posted, last_edited_at`

	err = tx.QueryRow(ctx, query,
		recipe.ID,
		recipe.Name,
		recipe.RecipeType,
		recipe.TimeToCook,
		recipe.Description,
	).Scan(&recipe.UserID, &recipe.DatePosted, &recipe.LastEditedAt)
	if err != nil {
		return nil, translatePostgresError(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM ingredients WHERE recipe_id = $1`, recipe.ID); err != nil {
		return nil, fmt.Errorf("clearing recipe ingredients: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM recipe_steps WHERE recipe_id = $1`, recipe.ID); err != nil {
		return nil, fmt.Errorf("clearing recipe steps: %w", err)
	}
	if err := replaceStructuredRecipeRows(ctx, tx, recipe); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing recipe update: %w", err)
	}
	return recipe, nil
}

func (r *RecipePostgres) Delete(ctx context.Context, id string) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE recipes
		SET deleted_at = now(), last_edited_at = now()
		WHERE recipe_id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return translatePostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRecipe(row rowScanner) (*Recipe, error) {
	var recipe Recipe
	err := row.Scan(
		&recipe.ID,
		&recipe.Name,
		&recipe.RecipeType,
		&recipe.TimeToCook,
		&recipe.Description,
		&recipe.UserID,
		&recipe.DatePosted,
		&recipe.LastEditedAt,
		&recipe.CoverImageID,
	)
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}

func replaceStructuredRecipeRows(ctx context.Context, tx pgx.Tx, recipe *Recipe) error {
	const ingredientQuery = `
		INSERT INTO ingredients (recipe_id, name, quantity, position)
		VALUES ($1, $2, $3, $4)
		RETURNING ingredient_id`
	for index := range recipe.Ingredients {
		ingredient := &recipe.Ingredients[index]
		ingredient.RecipeID = recipe.ID
		ingredient.Position = index
		if err := tx.QueryRow(ctx, ingredientQuery,
			recipe.ID,
			ingredient.Name,
			ingredient.Quantity,
			ingredient.Position,
		).Scan(&ingredient.ID); err != nil {
			return fmt.Errorf("inserting recipe ingredient: %w", translatePostgresError(err))
		}
	}

	const stepQuery = `
		INSERT INTO recipe_steps (recipe_id, step_number, instruction)
		VALUES ($1, $2, $3)
		RETURNING step_id`
	for index := range recipe.Steps {
		step := &recipe.Steps[index]
		step.RecipeID = recipe.ID
		step.StepNumber = index + 1
		if err := tx.QueryRow(ctx, stepQuery,
			recipe.ID,
			step.StepNumber,
			step.Instruction,
		).Scan(&step.ID); err != nil {
			return fmt.Errorf("inserting recipe step: %w", translatePostgresError(err))
		}
	}

	return nil
}

func (r *RecipePostgres) loadRecipeChildren(ctx context.Context, recipe *Recipe) error {
	ingredients, err := r.loadIngredients(ctx, recipe.ID)
	if err != nil {
		return err
	}
	steps, err := r.loadSteps(ctx, recipe.ID)
	if err != nil {
		return err
	}
	images, err := r.loadImages(ctx, recipe.ID)
	if err != nil {
		return err
	}

	recipe.Ingredients = ingredients
	recipe.Steps = steps
	recipe.Images = images
	return nil
}

func (r *RecipePostgres) loadIngredients(ctx context.Context, recipeID string) ([]Ingredient, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ingredient_id, recipe_id, name, quantity, position
		FROM ingredients
		WHERE recipe_id = $1
		ORDER BY position`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("loading recipe ingredients: %w", err)
	}
	defer rows.Close()

	ingredients := make([]Ingredient, 0)
	for rows.Next() {
		var ingredient Ingredient
		if err := rows.Scan(&ingredient.ID, &ingredient.RecipeID, &ingredient.Name, &ingredient.Quantity, &ingredient.Position); err != nil {
			return nil, fmt.Errorf("scanning recipe ingredient: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recipe ingredients: %w", err)
	}
	return ingredients, nil
}

func (r *RecipePostgres) loadSteps(ctx context.Context, recipeID string) ([]RecipeStep, error) {
	rows, err := r.db.Query(ctx, `
		SELECT step_id, recipe_id, step_number, instruction
		FROM recipe_steps
		WHERE recipe_id = $1
		ORDER BY step_number`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("loading recipe steps: %w", err)
	}
	defer rows.Close()

	steps := make([]RecipeStep, 0)
	for rows.Next() {
		var step RecipeStep
		if err := rows.Scan(&step.ID, &step.RecipeID, &step.StepNumber, &step.Instruction); err != nil {
			return nil, fmt.Errorf("scanning recipe step: %w", err)
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recipe steps: %w", err)
	}
	return steps, nil
}

func (r *RecipePostgres) loadImages(ctx context.Context, recipeID string) ([]RecipeImage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT image_id, recipe_id, s3_key, file_name, content_type,
		       file_size, position, is_cover, uploaded_at
		FROM recipe_images
		WHERE recipe_id = $1
		ORDER BY position`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("loading recipe images: %w", err)
	}
	defer rows.Close()

	images := make([]RecipeImage, 0)
	for rows.Next() {
		var image RecipeImage
		if err := rows.Scan(
			&image.ID,
			&image.RecipeID,
			&image.S3Key,
			&image.FileName,
			&image.ContentType,
			&image.FileSize,
			&image.Position,
			&image.IsCover,
			&image.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning recipe image: %w", err)
		}
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recipe images: %w", err)
	}
	return images, nil
}

func (r *RecipePostgres) AddImage(ctx context.Context, recipeID string, image *RecipeImage) (*RecipeImage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning image insert: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck -- rollback after commit is harmless

	var position int
	var isCover bool
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(i.position), -1) + 1, COUNT(i.image_id) = 0
		FROM recipes r
		LEFT JOIN recipe_images i ON i.recipe_id = r.recipe_id
		WHERE r.recipe_id = $1 AND r.deleted_at IS NULL AND r.recipe_type = 'image'
		GROUP BY r.recipe_id`, recipeID).Scan(&position, &isCover)
	if err != nil {
		return nil, translatePostgresError(err)
	}

	image.RecipeID = recipeID
	image.Position = position
	image.IsCover = isCover
	err = tx.QueryRow(ctx, `
		INSERT INTO recipe_images
			(recipe_id, s3_key, file_name, content_type, file_size, position, is_cover)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING image_id, uploaded_at`,
		recipeID, image.S3Key, image.FileName, image.ContentType, image.FileSize, position, isCover,
	).Scan(&image.ID, &image.UploadedAt)
	if err != nil {
		return nil, translatePostgresError(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE recipes SET last_edited_at = now() WHERE recipe_id = $1`, recipeID); err != nil {
		return nil, fmt.Errorf("touching recipe after image insert: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing image insert: %w", err)
	}
	return image, nil
}

func (r *RecipePostgres) GetImage(ctx context.Context, imageID string) (*RecipeImage, error) {
	var image RecipeImage
	err := r.db.QueryRow(ctx, `
		SELECT i.image_id, i.recipe_id, i.s3_key, i.file_name, i.content_type,
		       i.file_size, i.position, i.is_cover, i.uploaded_at
		FROM recipe_images i
		JOIN recipes r ON r.recipe_id = i.recipe_id
		WHERE i.image_id = $1 AND r.deleted_at IS NULL`, imageID).Scan(
		&image.ID, &image.RecipeID, &image.S3Key, &image.FileName, &image.ContentType,
		&image.FileSize, &image.Position, &image.IsCover, &image.UploadedAt,
	)
	if err != nil {
		return nil, translatePostgresError(err)
	}
	return &image, nil
}

func (r *RecipePostgres) DeleteImage(ctx context.Context, recipeID, imageID string) (*RecipeImage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning image delete: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck -- rollback after commit is harmless

	var image RecipeImage
	err = tx.QueryRow(ctx, `
		DELETE FROM recipe_images i
		USING recipes r
		WHERE i.image_id = $1 AND i.recipe_id = $2
		  AND r.recipe_id = i.recipe_id AND r.deleted_at IS NULL
		RETURNING i.image_id, i.recipe_id, i.s3_key, i.file_name, i.content_type,
		          i.file_size, i.position, i.is_cover, i.uploaded_at`, imageID, recipeID).Scan(
		&image.ID, &image.RecipeID, &image.S3Key, &image.FileName, &image.ContentType,
		&image.FileSize, &image.Position, &image.IsCover, &image.UploadedAt,
	)
	if err != nil {
		return nil, translatePostgresError(err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE recipe_images SET position = position + 1000000
		WHERE recipe_id = $1 AND position > $2`, recipeID, image.Position); err != nil {
		return nil, fmt.Errorf("reordering images after delete: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE recipe_images SET position = position - 1000001
		WHERE recipe_id = $1 AND position >= 1000000`, recipeID); err != nil {
		return nil, fmt.Errorf("finalizing image order after delete: %w", err)
	}
	if image.IsCover {
		if _, err := tx.Exec(ctx, `
			UPDATE recipe_images SET is_cover = true
			WHERE image_id = (SELECT image_id FROM recipe_images WHERE recipe_id = $1 ORDER BY position LIMIT 1)`, recipeID); err != nil {
			return nil, fmt.Errorf("selecting replacement cover: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE recipes SET last_edited_at = now() WHERE recipe_id = $1`, recipeID); err != nil {
		return nil, fmt.Errorf("touching recipe after image delete: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing image delete: %w", err)
	}
	return &image, nil
}

func (r *RecipePostgres) SetCoverImage(ctx context.Context, recipeID, imageID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning cover update: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck -- rollback after commit is harmless

	if _, err := tx.Exec(ctx, `UPDATE recipe_images SET is_cover = false WHERE recipe_id = $1`, recipeID); err != nil {
		return fmt.Errorf("clearing recipe cover: %w", err)
	}
	result, err := tx.Exec(ctx, `
		UPDATE recipe_images i SET is_cover = true
		FROM recipes r
		WHERE i.image_id = $1 AND i.recipe_id = $2
		  AND r.recipe_id = i.recipe_id AND r.deleted_at IS NULL`, imageID, recipeID)
	if err != nil {
		return translatePostgresError(err)
	}
	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `UPDATE recipes SET last_edited_at = now() WHERE recipe_id = $1`, recipeID); err != nil {
		return fmt.Errorf("touching recipe after cover update: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing cover update: %w", err)
	}
	return nil
}
