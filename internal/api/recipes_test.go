package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/repository"
)

const (
	testRecipeID = "11111111-1111-4111-8111-111111111111"
	otherUserID  = "22222222-2222-4222-8222-222222222222"
)

func TestRecipeRoutesRequireAuthentication(t *testing.T) {
	configureTestAuth(t)
	router := NewRouter(fakeUserRepository{}, fakeRecipeRepository{})

	for _, test := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/recipes"},
		{method: http.MethodPost, path: "/api/recipes"},
		{method: http.MethodGet, path: "/api/recipes/" + testRecipeID},
		{method: http.MethodPut, path: "/api/recipes/" + testRecipeID},
		{method: http.MethodDelete, path: "/api/recipes/" + testRecipeID},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assertErrorResponse(t, rec, http.StatusUnauthorized, "unauthorized")
		})
	}
}

func TestCreateStructuredRecipe(t *testing.T) {
	configureTestAuth(t)
	joined := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	recipes := fakeRecipeRepository{
		create: func(ctx context.Context, recipe *repository.Recipe) (*repository.Recipe, error) {
			if recipe.UserID != "viewer-1" {
				t.Fatalf("owner = %q, want viewer-1", recipe.UserID)
			}
			if recipe.Name != "Tomato Soup" || recipe.Ingredients[0].Name != "Tomatoes" {
				t.Fatalf("recipe was not normalized: %#v", recipe)
			}
			if recipe.Ingredients[0].Position != 0 || recipe.Steps[0].StepNumber != 1 {
				t.Fatalf("child ordering was not assigned: %#v", recipe)
			}
			recipe.ID = testRecipeID
			recipe.DatePosted = joined
			recipe.LastEditedAt = joined
			return recipe, nil
		},
	}
	router := newAuthenticatedRecipeRouter(recipes, "user")
	body := `{
		"name":"  Tomato Soup  ",
		"recipeType":"structured",
		"timeToCook":"30 minutes",
		"description":"Easy soup",
		"ingredients":[{"name":" Tomatoes ","quantity":"4"}],
		"steps":[{"instruction":"Simmer."}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader(body))
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if location := rec.Header().Get("Location"); location != "/api/recipes/"+testRecipeID {
		t.Fatalf("Location = %q", location)
	}
	var response recipeResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if response.ID != testRecipeID || len(response.Ingredients) != 1 || len(response.Steps) != 1 {
		t.Fatalf("response = %#v", response)
	}
}

func TestCreateRecipeValidation(t *testing.T) {
	configureTestAuth(t)
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{
			name:     "unknown field",
			body:     `{"name":"Soup","recipeType":"structured","ingredients":[],"steps":[],"surprise":true}`,
			wantCode: "invalid_request",
		},
		{
			name:     "blank name",
			body:     `{"name":" ","recipeType":"structured","ingredients":[{"name":"Salt"}],"steps":[{"instruction":"Mix"}]}`,
			wantCode: "invalid_name",
		},
		{
			name:     "missing ingredients",
			body:     `{"name":"Soup","recipeType":"structured","ingredients":[],"steps":[{"instruction":"Mix"}]}`,
			wantCode: "invalid_ingredients",
		},
		{
			name:     "missing steps",
			body:     `{"name":"Soup","recipeType":"structured","ingredients":[{"name":"Salt"}],"steps":[]}`,
			wantCode: "invalid_steps",
		},
		{
			name:     "image recipe waits for upload flow",
			body:     `{"name":"Grandma's card","recipeType":"image"}`,
			wantCode: "image_upload_not_supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newAuthenticatedRecipeRouter(fakeRecipeRepository{}, "user")
			req := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader(tt.body))
			addTestSession(t, req, "viewer-1")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assertErrorResponse(t, rec, http.StatusBadRequest, tt.wantCode)
		})
	}
}

func TestListRecipes(t *testing.T) {
	configureTestAuth(t)
	recipes := fakeRecipeRepository{
		list: func(ctx context.Context) ([]*repository.Recipe, error) {
			return []*repository.Recipe{{
				ID: testRecipeID, Name: "Soup", RecipeType: "structured", UserID: "viewer-1",
			}}, nil
		},
	}
	router := newAuthenticatedRecipeRouter(recipes, "user")
	req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var response []recipeSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(response) != 1 || response[0].Name != "Soup" {
		t.Fatalf("response = %#v", response)
	}
}

func TestUpdateRecipeAuthorization(t *testing.T) {
	configureTestAuth(t)
	body := `{"name":"Soup","recipeType":"structured","ingredients":[{"name":"Salt"}],"steps":[{"instruction":"Mix"}]}`

	tests := []struct {
		name       string
		role       string
		ownerID    string
		wantStatus int
	}{
		{name: "owner", role: "user", ownerID: "viewer-1", wantStatus: http.StatusOK},
		{name: "admin", role: "admin", ownerID: otherUserID, wantStatus: http.StatusOK},
		{name: "other user", role: "user", ownerID: otherUserID, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			recipes := fakeRecipeRepository{
				getByID: func(ctx context.Context, id string) (*repository.Recipe, error) {
					return &repository.Recipe{ID: id, UserID: tt.ownerID}, nil
				},
				update: func(ctx context.Context, recipe *repository.Recipe) (*repository.Recipe, error) {
					updateCalled = true
					recipe.UserID = tt.ownerID
					return recipe, nil
				},
			}
			router := newAuthenticatedRecipeRouter(recipes, tt.role)
			req := httptest.NewRequest(http.MethodPut, "/api/recipes/"+testRecipeID, strings.NewReader(body))
			addTestSession(t, req, "viewer-1")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if updateCalled != (tt.wantStatus == http.StatusOK) {
				t.Fatalf("update called = %v", updateCalled)
			}
		})
	}
}

func TestDeleteRecipe(t *testing.T) {
	configureTestAuth(t)
	deletedID := ""
	recipes := fakeRecipeRepository{
		getByID: func(ctx context.Context, id string) (*repository.Recipe, error) {
			return &repository.Recipe{ID: id, UserID: "viewer-1"}, nil
		},
		delete: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
	}
	router := newAuthenticatedRecipeRouter(recipes, "user")
	req := httptest.NewRequest(http.MethodDelete, "/api/recipes/"+testRecipeID, nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if deletedID != testRecipeID {
		t.Fatalf("deleted ID = %q", deletedID)
	}
}

func TestRecipeNotFound(t *testing.T) {
	configureTestAuth(t)
	recipes := fakeRecipeRepository{
		getByID: func(ctx context.Context, id string) (*repository.Recipe, error) {
			return nil, apperror.ErrNotFound
		},
	}
	router := newAuthenticatedRecipeRouter(recipes, "user")
	req := httptest.NewRequest(http.MethodGet, "/api/recipes/"+testRecipeID, nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, "not_found")
}

func TestListRecipesRepositoryFailure(t *testing.T) {
	configureTestAuth(t)
	recipes := fakeRecipeRepository{
		list: func(ctx context.Context) ([]*repository.Recipe, error) {
			return nil, errors.New("db down")
		},
	}
	router := newAuthenticatedRecipeRouter(recipes, "user")
	req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusInternalServerError, "internal_error")
}

func newAuthenticatedRecipeRouter(recipes repository.RecipeRepository, role string) http.Handler {
	users := fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return &repository.User{ID: id, Email: testViewerEmail, Alias: "Viewer", Role: role}, nil
		},
	}
	return NewRouter(users, recipes)
}
