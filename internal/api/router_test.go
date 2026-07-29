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

	"github.com/gorilla/securecookie"
)

const (
	testSessionSecret = "01234567890123456789012345678901"
	testViewerEmail   = "viewer@example.com"
	testUserID        = "33333333-3333-4333-8333-333333333333"
	missingUserID     = "44444444-4444-4444-8444-444444444444"
)

type fakeUserRepository struct {
	create              func(ctx context.Context, u *repository.User) (*repository.User, error)
	getByID             func(ctx context.Context, id string) (*repository.User, error)
	getByProviderUserID func(ctx context.Context, provider, providerUserID string) (*repository.User, error)
}

type fakeRecipeRepository struct {
	create        func(context.Context, *repository.Recipe) (*repository.Recipe, error)
	list          func(context.Context) ([]*repository.Recipe, error)
	getByID       func(context.Context, string) (*repository.Recipe, error)
	update        func(context.Context, *repository.Recipe) (*repository.Recipe, error)
	delete        func(context.Context, string) error
	addImage      func(context.Context, string, *repository.RecipeImage) (*repository.RecipeImage, error)
	getImage      func(context.Context, string) (*repository.RecipeImage, error)
	deleteImage   func(context.Context, string, string) (*repository.RecipeImage, error)
	setCoverImage func(context.Context, string, string) error
}

func (f fakeRecipeRepository) AddImage(ctx context.Context, recipeID string, image *repository.RecipeImage) (*repository.RecipeImage, error) {
	if f.addImage == nil {
		return nil, errors.New("not implemented")
	}
	return f.addImage(ctx, recipeID, image)
}

func (f fakeRecipeRepository) GetImage(ctx context.Context, imageID string) (*repository.RecipeImage, error) {
	if f.getImage == nil {
		return nil, errors.New("not implemented")
	}
	return f.getImage(ctx, imageID)
}

func (f fakeRecipeRepository) DeleteImage(ctx context.Context, recipeID, imageID string) (*repository.RecipeImage, error) {
	if f.deleteImage == nil {
		return nil, errors.New("not implemented")
	}
	return f.deleteImage(ctx, recipeID, imageID)
}

func (f fakeRecipeRepository) SetCoverImage(ctx context.Context, recipeID, imageID string) error {
	if f.setCoverImage == nil {
		return errors.New("not implemented")
	}
	return f.setCoverImage(ctx, recipeID, imageID)
}

type fakeHealthChecker struct {
	err error
}

func (f fakeHealthChecker) Ping(context.Context) error {
	return f.err
}

func (f fakeRecipeRepository) Create(ctx context.Context, recipe *repository.Recipe) (*repository.Recipe, error) {
	if f.create == nil {
		return nil, errors.New("not implemented")
	}
	return f.create(ctx, recipe)
}

func (f fakeRecipeRepository) List(ctx context.Context) ([]*repository.Recipe, error) {
	if f.list == nil {
		return nil, errors.New("not implemented")
	}
	return f.list(ctx)
}

func (f fakeRecipeRepository) GetByID(ctx context.Context, id string) (*repository.Recipe, error) {
	if f.getByID == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByID(ctx, id)
}

func (f fakeRecipeRepository) Update(ctx context.Context, recipe *repository.Recipe) (*repository.Recipe, error) {
	if f.update == nil {
		return nil, errors.New("not implemented")
	}
	return f.update(ctx, recipe)
}

func (f fakeRecipeRepository) Delete(ctx context.Context, id string) error {
	if f.delete == nil {
		return errors.New("not implemented")
	}
	return f.delete(ctx, id)
}

func newTestRouter(users repository.UserRepository) http.Handler {
	return NewRouter(users, fakeRecipeRepository{}, fakeHealthChecker{}, nil)
}

func (f fakeUserRepository) Create(ctx context.Context, u *repository.User) (*repository.User, error) {
	if f.create == nil {
		return nil, errors.New("not implemented")
	}
	return f.create(ctx, u)
}

func (f fakeUserRepository) GetByID(ctx context.Context, id string) (*repository.User, error) {
	if f.getByID == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByID(ctx, id)
}

func (f fakeUserRepository) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*repository.User, error) {
	if f.getByProviderUserID == nil {
		return nil, errors.New("not implemented")
	}
	return f.getByProviderUserID(ctx, provider, providerUserID)
}

func TestHealth(t *testing.T) {
	router := newTestRouter(fakeUserRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want ok", body["status"])
	}
}

func TestReadiness(t *testing.T) {
	for _, tt := range []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{name: "database available", wantStatus: http.StatusOK, wantBody: `{"status":"ready"}`},
		{name: "database unavailable", err: errors.New("db down"), wantStatus: http.StatusServiceUnavailable, wantBody: `{"error":"not_ready"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(fakeUserRepository{}, fakeRecipeRepository{}, fakeHealthChecker{err: tt.err}, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if body := strings.TrimSpace(rec.Body.String()); body != tt.wantBody {
				t.Fatalf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	configureTestAuth(t)
	joined := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)

	router := newTestRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			if id == "viewer-1" {
				return testViewer(), nil
			}
			if id != testUserID {
				t.Fatalf("id = %q, want %s", id, testUserID)
			}

			return &repository.User{
				ID:             testUserID,
				Email:          "drew@example.com",
				Provider:       "google",
				ProviderUserID: "google-123",
				Alias:          "Drew",
				Role:           "user",
				DateJoined:     joined,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+testUserID, nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body userSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if body.ID != testUserID {
		t.Fatalf("id = %q, want %s", body.ID, testUserID)
	}
	if body.Alias != "Drew" {
		t.Fatalf("alias = %q, want Drew", body.Alias)
	}
	if !body.DateJoined.Equal(joined) {
		t.Fatalf("dateJoined = %v, want %v", body.DateJoined, joined)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	configureTestAuth(t)
	router := newTestRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			if id == "viewer-1" {
				return testViewer(), nil
			}
			return nil, apperror.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+missingUserID, nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, "not_found")
}

func TestGetUserByIDInternalError(t *testing.T) {
	configureTestAuth(t)
	router := newTestRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			if id == "viewer-1" {
				return testViewer(), nil
			}
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+testUserID, nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusInternalServerError, "internal_error")
}

func TestGetUserByIDRejectsInvalidID(t *testing.T) {
	configureTestAuth(t)
	router := newTestRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			if id == "viewer-1" {
				return testViewer(), nil
			}
			t.Fatal("repository should not be called for an invalid ID")
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/not-a-uuid", nil)
	addTestSession(t, req, "viewer-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusBadRequest, "invalid_id")
}

func TestGetUserByIDRequiresAuthentication(t *testing.T) {
	configureTestAuth(t)
	router := newTestRouter(fakeUserRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+testUserID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestGetUserByIDUnsupportedMethod(t *testing.T) {
	router := newTestRouter(fakeUserRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/users/"+testUserID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCurrentUserNotConfigured(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("ALLOWED_EMAILS", "")

	router := newTestRouter(fakeUserRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusServiceUnavailable, "auth_not_configured")
}

func configureTestAuth(t *testing.T) {
	t.Helper()
	t.Setenv("GOOGLE_CLIENT_ID", "test-client")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	t.Setenv("SESSION_SECRET", testSessionSecret)
	t.Setenv("ALLOWED_EMAILS", testViewerEmail)
}

func addTestSession(t *testing.T, req *http.Request, userID string) {
	t.Helper()

	codec := securecookie.New([]byte(testSessionSecret), nil).MaxAge(int(defaultSessionTTL.Seconds()))
	value, err := codec.Encode(defaultSessionCookieName, sessionPayload{
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("encoding test session: %v", err)
	}

	req.AddCookie(&http.Cookie{Name: defaultSessionCookieName, Value: value})
}

func testViewer() *repository.User {
	return &repository.User{
		ID:    "viewer-1",
		Email: testViewerEmail,
		Alias: "Viewer",
		Role:  "user",
	}
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if body.Error != wantCode {
		t.Fatalf("error = %q, want %q", body.Error, wantCode)
	}
}
