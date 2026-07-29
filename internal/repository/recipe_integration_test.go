package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func TestRecipePostgresLifecycle(t *testing.T) {
	if os.Getenv("RUN_DATABASE_INTEGRATION") == "" {
		t.Skip("set RUN_DATABASE_INTEGRATION=1 to run against local Postgres")
	}
	if err := godotenv.Load("../../.env"); err != nil && os.Getenv("DB_HOST") == "" {
		t.Fatalf("loading test database environment: %v", err)
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = os.Getenv("DB_PORT")
	}
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("pinging test database: %v", err)
	}
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("applying test migrations: %v", err)
	}

	users := NewUserPostgres(pool)
	recipes := NewRecipePostgres(pool)
	unique := time.Now().UnixNano()
	user, err := users.Create(ctx, &User{
		Email:          fmt.Sprintf("recipe-integration-%d@example.com", unique),
		Provider:       "google",
		ProviderUserID: fmt.Sprintf("recipe-integration-%d", unique),
		Alias:          "Integration Test",
	})
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	defer func() {
		if _, err := pool.Exec(ctx, `DELETE FROM recipes WHERE user_id = $1`, user.ID); err != nil {
			t.Errorf("cleaning up test recipes: %v", err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM users WHERE user_id = $1`, user.ID); err != nil {
			t.Errorf("cleaning up test user: %v", err)
		}
	}()

	created, err := recipes.Create(ctx, &Recipe{
		Name:        "Integration Soup",
		RecipeType:  "structured",
		TimeToCook:  "20 minutes",
		Description: "Repository lifecycle test",
		UserID:      user.ID,
		Ingredients: []Ingredient{{Name: "Tomatoes", Quantity: "4"}},
		Steps:       []RecipeStep{{Instruction: "Simmer."}},
	})
	if err != nil {
		t.Fatalf("creating recipe: %v", err)
	}
	if created.ID == "" || created.Ingredients[0].ID == "" || created.Steps[0].ID == "" {
		t.Fatalf("created recipe is missing generated IDs: %#v", created)
	}

	loaded, err := recipes.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("loading recipe: %v", err)
	}
	if len(loaded.Ingredients) != 1 || len(loaded.Steps) != 1 || loaded.Ingredients[0].Name != "Tomatoes" {
		t.Fatalf("loaded recipe children = %#v", loaded)
	}

	firstImage, err := recipes.AddImage(ctx, created.ID, &RecipeImage{
		S3Key: fmt.Sprintf("integration-%d-one.jpg", unique), FileName: "one.jpg", ContentType: "image/jpeg", FileSize: 100,
	})
	if err != nil {
		t.Fatalf("adding first recipe image: %v", err)
	}
	secondImage, err := recipes.AddImage(ctx, created.ID, &RecipeImage{
		S3Key: fmt.Sprintf("integration-%d-two.jpg", unique), FileName: "two.jpg", ContentType: "image/jpeg", FileSize: 200,
	})
	if err != nil {
		t.Fatalf("adding second recipe image: %v", err)
	}
	if !firstImage.IsCover || secondImage.IsCover || secondImage.Position != 1 {
		t.Fatalf("initial image ordering = %#v, %#v", firstImage, secondImage)
	}
	if err := recipes.SetCoverImage(ctx, created.ID, secondImage.ID); err != nil {
		t.Fatalf("setting recipe cover: %v", err)
	}
	if _, err := recipes.DeleteImage(ctx, created.ID, firstImage.ID); err != nil {
		t.Fatalf("deleting first recipe image: %v", err)
	}
	remainingImage, err := recipes.GetImage(ctx, secondImage.ID)
	if err != nil || remainingImage.Position != 0 || !remainingImage.IsCover {
		t.Fatalf("remaining image = %#v, error = %v", remainingImage, err)
	}

	loaded.Name = "Updated Integration Soup"
	loaded.Ingredients = append(loaded.Ingredients, Ingredient{Name: "Salt", Quantity: "to taste"})
	loaded.Steps = []RecipeStep{{Instruction: "Combine."}, {Instruction: "Simmer."}}
	updated, err := recipes.Update(ctx, loaded)
	if err != nil {
		t.Fatalf("updating recipe: %v", err)
	}
	if updated.Name != "Updated Integration Soup" || len(updated.Ingredients) != 2 || updated.Steps[1].StepNumber != 2 {
		t.Fatalf("updated recipe = %#v", updated)
	}

	listed, err := recipes.List(ctx)
	if err != nil {
		t.Fatalf("listing recipes: %v", err)
	}
	found := false
	for _, recipe := range listed {
		if recipe.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created recipe was not returned by List")
	}

	if err := recipes.Delete(ctx, created.ID); err != nil {
		t.Fatalf("deleting recipe: %v", err)
	}
	if _, err := recipes.GetByID(ctx, created.ID); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want not found", err)
	}
}
