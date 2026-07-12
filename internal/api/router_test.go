package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/repository"
)

type fakeUserRepository struct {
	create              func(ctx context.Context, u *repository.User) (*repository.User, error)
	getByID             func(ctx context.Context, id string) (*repository.User, error)
	getByProviderUserID func(ctx context.Context, provider, providerUserID string) (*repository.User, error)
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
	router := NewRouter(fakeUserRepository{})

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

func TestGetUserByID(t *testing.T) {
	joined := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)

	router := NewRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			if id != "user-1" {
				t.Fatalf("id = %q, want user-1", id)
			}

			return &repository.User{
				ID:             "user-1",
				Email:          "drew@example.com",
				Provider:       "google",
				ProviderUserID: "google-123",
				Alias:          "Drew",
				Role:           "user",
				DateJoined:     joined,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body userResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if body.ID != "user-1" {
		t.Fatalf("id = %q, want user-1", body.ID)
	}
	if body.Email != "drew@example.com" {
		t.Fatalf("email = %q, want drew@example.com", body.Email)
	}
	if !body.DateJoined.Equal(joined) {
		t.Fatalf("dateJoined = %v, want %v", body.DateJoined, joined)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	router := NewRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return nil, apperror.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, "not_found")
}

func TestGetUserByIDInternalError(t *testing.T) {
	router := NewRouter(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusInternalServerError, "internal_error")
}

func TestGetUserByIDUnsupportedMethod(t *testing.T) {
	router := NewRouter(fakeUserRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/users/user-1", nil)
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

	router := NewRouter(fakeUserRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusServiceUnavailable, "auth_not_configured")
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
