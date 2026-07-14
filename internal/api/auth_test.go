package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/repository"
)

func TestAuthConfigurationFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*authConfig)
	}{
		{
			name: "empty allowlist",
			mutate: func(cfg *authConfig) {
				cfg.AllowedEmails = nil
			},
		},
		{
			name: "short session secret",
			mutate: func(cfg *authConfig) {
				cfg.SessionSecret = "too-short"
			},
		},
		{
			name: "missing google client",
			mutate: func(cfg *authConfig) {
				cfg.GoogleClientID = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testAuthConfig()
			tt.mutate(&cfg)
			handler := newAuthHandler(fakeUserRepository{}, cfg)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/google/login", nil)
			rec := httptest.NewRecorder()
			handler.Login(rec, req)

			assertErrorResponse(t, rec, http.StatusServiceUnavailable, "auth_not_configured")
		})
	}
}

func TestSessionValidationDoesNotDependOnGoogleConfiguration(t *testing.T) {
	cfg := testAuthConfig()
	cfg.GoogleClientID = ""
	cfg.GoogleClientSecret = ""
	handler := newAuthHandler(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return testViewer(), nil
		},
	}, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	addSessionForHandler(t, handler, req, "viewer-1", time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoginSetsProtectedStateCookie(t *testing.T) {
	cfg := testAuthConfig()
	cfg.SessionCookieSecure = true
	handler := newAuthHandler(fakeUserRepository{}, cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/login", nil)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get("Location"); !strings.HasPrefix(location, "https://accounts.google.com/o/oauth2/auth?") {
		t.Fatalf("redirect location = %q, want Google authorization URL", location)
	}

	cookie := responseCookie(t, rec, stateCookieName)
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("state cookie flags = HttpOnly:%v Secure:%v SameSite:%v", cookie.HttpOnly, cookie.Secure, cookie.SameSite)
	}
	if cookie.Path != "/api/auth/google" {
		t.Fatalf("state cookie path = %q", cookie.Path)
	}
}

func TestVerifyState(t *testing.T) {
	handler := newAuthHandler(fakeUserRepository{}, testAuthConfig())

	tests := []struct {
		name       string
		queryState string
		payload    statePayload
		tamper     bool
		wantErr    bool
	}{
		{
			name:       "valid",
			queryState: "expected",
			payload:    statePayload{State: "expected", ExpiresAt: time.Now().Add(time.Minute).Unix()},
		},
		{
			name:       "mismatch",
			queryState: "unexpected",
			payload:    statePayload{State: "expected", ExpiresAt: time.Now().Add(time.Minute).Unix()},
			wantErr:    true,
		},
		{
			name:       "expired",
			queryState: "expected",
			payload:    statePayload{State: "expected", ExpiresAt: time.Now().Add(-time.Minute).Unix()},
			wantErr:    true,
		},
		{
			name:       "tampered cookie",
			queryState: "expected",
			payload:    statePayload{State: "expected", ExpiresAt: time.Now().Add(time.Minute).Unix()},
			tamper:     true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := handler.codec.Encode(stateCookieName, tt.payload)
			if err != nil {
				t.Fatalf("encoding state: %v", err)
			}
			if tt.tamper {
				value += "tampered"
			}

			req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state="+tt.queryState, nil)
			req.AddCookie(&http.Cookie{Name: stateCookieName, Value: value})
			err = handler.verifyState(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("verifyState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeRejectsInvalidSessions(t *testing.T) {
	handler := newAuthHandler(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return testViewer(), nil
		},
	}, testAuthConfig())

	tests := []struct {
		name      string
		addCookie func(*testing.T, *http.Request)
	}{
		{name: "missing cookie"},
		{
			name: "tampered cookie",
			addCookie: func(t *testing.T, req *http.Request) {
				value, err := handler.codec.Encode(handler.cfg.SessionCookieName, sessionPayload{
					UserID:    "viewer-1",
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
				})
				if err != nil {
					t.Fatalf("encoding test session: %v", err)
				}
				req.AddCookie(&http.Cookie{Name: handler.cfg.SessionCookieName, Value: value + "tampered"})
			},
		},
		{
			name: "expired payload",
			addCookie: func(t *testing.T, req *http.Request) {
				addSessionForHandler(t, handler, req, "viewer-1", time.Now().Add(-time.Minute))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			if tt.addCookie != nil {
				tt.addCookie(t, req)
			}
			rec := httptest.NewRecorder()

			handler.Me(rec, req)

			assertErrorResponse(t, rec, http.StatusUnauthorized, "unauthorized")
		})
	}
}

func TestMeRejectsUserRemovedFromAllowlist(t *testing.T) {
	cfg := testAuthConfig()
	cfg.AllowedEmails = parseAllowedEmails("someone-else@example.com")
	handler := newAuthHandler(fakeUserRepository{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return testViewer(), nil
		},
	}, cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	addSessionForHandler(t, handler, req, "viewer-1", time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	assertErrorResponse(t, rec, http.StatusUnauthorized, "unauthorized")
	cleared := responseCookie(t, rec, cfg.SessionCookieName)
	if cleared.MaxAge != -1 {
		t.Fatalf("cleared cookie MaxAge = %d, want -1", cleared.MaxAge)
	}
}

func TestSetSessionCookieUsesConfiguredSecurityAndTTL(t *testing.T) {
	cfg := testAuthConfig()
	cfg.SessionCookieSecure = true
	cfg.SessionTTL = 45 * 24 * time.Hour
	handler := newAuthHandler(fakeUserRepository{}, cfg)
	rec := httptest.NewRecorder()

	if err := handler.setSessionCookie(rec, testViewer()); err != nil {
		t.Fatalf("setSessionCookie() error = %v", err)
	}

	cookie := responseCookie(t, rec, cfg.SessionCookieName)
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("session cookie flags = HttpOnly:%v Secure:%v SameSite:%v", cookie.HttpOnly, cookie.Secure, cookie.SameSite)
	}
	if cookie.MaxAge != int(cfg.SessionTTL.Seconds()) {
		t.Fatalf("session cookie MaxAge = %d, want %d", cookie.MaxAge, int(cfg.SessionTTL.Seconds()))
	}

	var session sessionPayload
	if err := handler.codec.Decode(cfg.SessionCookieName, cookie.Value, &session); err != nil {
		t.Fatalf("decoding session cookie: %v", err)
	}
	if session.UserID != "viewer-1" {
		t.Fatalf("session user ID = %q, want viewer-1", session.UserID)
	}
}

func TestFindOrCreateUserRecoversFromConcurrentInsert(t *testing.T) {
	lookupCount := 0
	winner := &repository.User{ID: "winner", Email: testViewerEmail}
	handler := newAuthHandler(fakeUserRepository{
		getByProviderUserID: func(ctx context.Context, provider, providerUserID string) (*repository.User, error) {
			lookupCount++
			if lookupCount == 1 {
				return nil, apperror.ErrNotFound
			}
			return winner, nil
		},
		create: func(ctx context.Context, u *repository.User) (*repository.User, error) {
			return nil, apperror.ErrConflict
		},
	}, testAuthConfig())

	user, err := handler.findOrCreateUser(context.Background(), googleUserInfo{
		Subject: "google-subject",
		Email:   testViewerEmail,
		Name:    "Viewer",
	})

	if err != nil {
		t.Fatalf("findOrCreateUser() error = %v", err)
	}
	if user != winner {
		t.Fatalf("findOrCreateUser() = %#v, want concurrent winner", user)
	}
}

func TestMeMapsRepositoryFailures(t *testing.T) {
	tests := []struct {
		name       string
		repoErr    error
		wantStatus int
	}{
		{name: "deleted user", repoErr: apperror.ErrNotFound, wantStatus: http.StatusUnauthorized},
		{name: "database failure", repoErr: errors.New("db down"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newAuthHandler(fakeUserRepository{
				getByID: func(ctx context.Context, id string) (*repository.User, error) {
					return nil, tt.repoErr
				},
			}, testAuthConfig())
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			addSessionForHandler(t, handler, req, "viewer-1", time.Now().Add(time.Hour))
			rec := httptest.NewRecorder()

			handler.Me(rec, req)

			wantCode := "unauthorized"
			if tt.wantStatus == http.StatusInternalServerError {
				wantCode = "internal_error"
			}
			assertErrorResponse(t, rec, tt.wantStatus, wantCode)
		})
	}
}

func testAuthConfig() authConfig {
	return authConfig{
		GoogleClientID:      "test-client",
		GoogleClientSecret:  "test-client-secret",
		GoogleRedirectURL:   defaultGoogleRedirectURL,
		FrontendURL:         "/api/me",
		SessionCookieName:   defaultSessionCookieName,
		SessionCookieSecure: false,
		SessionTTL:          defaultSessionTTL,
		SessionSecret:       testSessionSecret,
		AllowedEmails:       parseAllowedEmails(testViewerEmail),
	}
}

func addSessionForHandler(t *testing.T, handler *AuthHandler, req *http.Request, userID string, expiresAt time.Time) {
	t.Helper()
	value, err := handler.codec.Encode(handler.cfg.SessionCookieName, sessionPayload{
		UserID:    userID,
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		t.Fatalf("encoding test session: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: handler.cfg.SessionCookieName, Value: value})
}

func responseCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response cookie %q not found", name)
	return nil
}
