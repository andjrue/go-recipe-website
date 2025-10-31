package repository

import (
	"context"
	"fmt"
	"recipe-website/internal/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresRecipeRepository struct {
	db *pgx.Conn
}

func (r *PostgresRecipeRepository) scanRecipe(row pgx.Row) (*domain.Recipe, error) {
	var recipe domain.Recipe

	err := row.Scan(
		&recipe.RecipeID,
		&recipe.DatePosted,
		&recipe.Description,
		&recipe.DatePosted,
		&recipe.LastEditedAt,
		&recipe.Name,
		&recipe.TimeToCook,
		&recipe.UserID,
	)

	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

func (r *PostgresRecipeRepository) GetByPK(ctx context.Context, recipeID string) (*domain.Recipe, error) {

	q := ` SELECT *
	FROM recipes
	WHERE recipe_id = $1 AND deleted = false
	`

	row := r.db.QueryRow(ctx, q, recipeID)
	recipe, err := r.scanRecipe(row)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("recipe not found: %v", err)
		}
		return nil, err
	}
	return recipe, nil
}

func (r *PostgresRecipeRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Recipe, error) {

	var recipes []domain.Recipe

	q := `SELECT *
	FROM recipes
	WHERE user_id = $1 AND deleted = false
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		recipe, err := r.scanRecipe(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %v", err)
		}

		recipes = append(recipes, *recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	return recipes, nil

}

func (r *PostgresRecipeRepository) Create(ctx context.Context, req *domain.CreateRecipeRequest) (*domain.Recipe, error) {
	q := `INSERT INTO recipes (name, time_to_cook, description, user_id, date_posted, deleted, last_edited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	now := time.Now()

	row, err := r.db.Query(ctx, q,
		req.Name,
		req.TimeToCook,
		req.Description,
		req.UserID,
		now,
		false,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe: %v", err)
	}

	defer row.Close()
	recipe, err := r.scanRecipe(row)
	if err != nil {
		return nil, fmt.Errorf("unable to scan created recipe: %v", err)
	}

	return recipe, nil
}

func (r *PostgresRecipeRepository) Update(ctx context.Context, req *domain.UpdateRecipeRequest) error {
	setClauses := []string{}
	args := []interface{}{}
	argPosition := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argPosition))
		args = append(args, *req.Name)
		argPosition++
	}

	if req.TimeToCook != nil {
		setClauses = append(setClauses, fmt.Sprintf("time_to_cook = $%d", argPosition))
		args = append(args, *req.TimeToCook)
		argPosition++
	}

	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argPosition))
		args = append(args, *req.Description)
		argPosition++
	}

	q := fmt.Sprintf(
		"UPDATE recipes SET %s WHERE id = %d",
		strings.Join(setClauses, ", "),
		argPosition,
	)
	args = append(args, req.RecipeID)

	commandTag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("not able to update recipe: %v", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no recipe updated")
	}

	return nil
}

func (r *PostgresRecipeRepository) Delete(ctx context.Context, recipeID string) error {
	return nil
}
