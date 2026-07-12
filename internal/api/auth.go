package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/repository"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	defaultGoogleRedirectURL = "http://localhost:8080/api/auth/google/callback"
	defaultSessionCookieName = "recipe_session"
	defaultSessionTTL        = 30 * 24 * time.Hour
	stateCookieName          = "recipe_oauth_state"
)

type AuthHandler struct {
	users repository.UserRepository
	cfg   authConfig
	codec *securecookie.SecureCookie
}

type authConfig struct {
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	FrontendURL         string
	SessionCookieName   string
	SessionCookieSecure bool
	SessionTTL          time.Duration
	SessionSecret       string
	AllowedEmails       map[string]struct{}
}

type sessionPayload struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expiresAt"`
}

type statePayload struct {
	State     string `json:"state"`
	ExpiresAt int64  `json:"expiresAt"`
}

type googleUserInfo struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

func NewAuthHandlerFromEnv(users repository.UserRepository) *AuthHandler {
	cfg := newAuthConfigFromEnv()
	if cfg.SessionSecret == "" {
		return &AuthHandler{users: users, cfg: cfg}
	}

	return &AuthHandler{
		users: users,
		cfg:   cfg,
		codec: securecookie.New([]byte(cfg.SessionSecret), nil),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeError(w, http.StatusServiceUnavailable, "auth_not_configured")
		return
	}

	state, err := randomToken(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if err := h.setStateCookie(w, state); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	oauthConfig := h.oauthConfig()
	http.Redirect(w, r, oauthConfig.AuthCodeURL(state), http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeError(w, http.StatusServiceUnavailable, "auth_not_configured")
		return
	}

	if errCode := r.URL.Query().Get("error"); errCode != "" {
		writeError(w, http.StatusUnauthorized, "oauth_denied")
		return
	}

	if err := h.verifyState(r); err != nil {
		h.clearStateCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid_oauth_state")
		return
	}
	h.clearStateCookie(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing_oauth_code")
		return
	}

	googleUser, err := h.exchangeGoogleUser(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_google_token")
		return
	}

	if !googleUser.EmailVerified {
		writeError(w, http.StatusForbidden, "email_not_verified")
		return
	}

	if !h.emailAllowed(googleUser.Email) {
		writeError(w, http.StatusForbidden, "email_not_allowed")
		return
	}

	user, err := h.findOrCreateUser(r.Context(), googleUser)
	if err != nil {
		if errors.Is(err, apperror.ErrConflict) {
			writeError(w, http.StatusConflict, "account_conflict")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if err := h.setSessionCookie(w, user); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	redirectURL := h.cfg.FrontendURL
	if redirectURL == "" {
		redirectURL = "/api/me"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(user))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) currentUser(w http.ResponseWriter, r *http.Request) (*repository.User, bool) {
	if !h.configured() {
		writeError(w, http.StatusServiceUnavailable, "auth_not_configured")
		return nil, false
	}

	cookie, err := r.Cookie(h.cfg.SessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, false
	}

	var session sessionPayload
	if err := h.codec.Decode(h.cfg.SessionCookieName, cookie.Value, &session); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, false
	}

	if session.UserID == "" || time.Now().After(time.Unix(session.ExpiresAt, 0)) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, false
	}

	user, err := h.users.GetByID(r.Context(), session.UserID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return nil, false
		}
		writeError(w, http.StatusInternalServerError, "internal_error")
		return nil, false
	}

	return user, true
}

func (h *AuthHandler) configured() bool {
	return h.users != nil &&
		h.codec != nil &&
		h.cfg.GoogleClientID != "" &&
		h.cfg.GoogleClientSecret != "" &&
		h.cfg.GoogleRedirectURL != "" &&
		h.cfg.SessionCookieName != "" &&
		h.cfg.SessionTTL > 0
}

func (h *AuthHandler) oauthConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     h.cfg.GoogleClientID,
		ClientSecret: h.cfg.GoogleClientSecret,
		RedirectURL:  h.cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (h *AuthHandler) exchangeGoogleUser(ctx context.Context, code string) (googleUserInfo, error) {
	oauthConfig := h.oauthConfig()
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return googleUserInfo{}, fmt.Errorf("exchanging google oauth code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return googleUserInfo{}, errors.New("google response missing id_token")
	}

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return googleUserInfo{}, fmt.Errorf("loading google oidc provider: %w", err)
	}

	idToken, err := provider.Verifier(&oidc.Config{ClientID: h.cfg.GoogleClientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return googleUserInfo{}, fmt.Errorf("validating google id token: %w", err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return googleUserInfo{}, fmt.Errorf("decoding google id token claims: %w", err)
	}

	if idToken.Subject == "" || claims.Email == "" {
		return googleUserInfo{}, errors.New("google token missing required claims")
	}

	return googleUserInfo{
		Subject:       idToken.Subject,
		Email:         strings.ToLower(claims.Email),
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
	}, nil
}

func (h *AuthHandler) findOrCreateUser(ctx context.Context, googleUser googleUserInfo) (*repository.User, error) {
	user, err := h.users.GetByProviderUserID(ctx, "google", googleUser.Subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	alias := googleUser.Name
	if alias == "" {
		alias = googleUser.Email
	}

	return h.users.Create(ctx, &repository.User{
		Email:          googleUser.Email,
		Provider:       "google",
		ProviderUserID: googleUser.Subject,
		Alias:          alias,
	})
}

func (h *AuthHandler) emailAllowed(email string) bool {
	if len(h.cfg.AllowedEmails) == 0 {
		return true
	}

	_, ok := h.cfg.AllowedEmails[strings.ToLower(email)]
	return ok
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, user *repository.User) error {
	session := sessionPayload{
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: time.Now().Add(h.cfg.SessionTTL).Unix(),
	}

	encoded, err := h.codec.Encode(h.cfg.SessionCookieName, session)
	if err != nil {
		return fmt.Errorf("encoding session cookie: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    encoded,
		Path:     "/",
		Expires:  time.Unix(session.ExpiresAt, 0),
		MaxAge:   int(h.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setStateCookie(w http.ResponseWriter, state string) error {
	payload := statePayload{
		State:     state,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
	}

	encoded, err := h.codec.Encode(stateCookieName, payload)
	if err != nil {
		return fmt.Errorf("encoding state cookie: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    encoded,
		Path:     "/api/auth/google",
		Expires:  time.Unix(payload.ExpiresAt, 0),
		MaxAge:   int((10 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (h *AuthHandler) verifyState(r *http.Request) error {
	queryState := r.URL.Query().Get("state")
	if queryState == "" {
		return errors.New("missing state")
	}

	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return fmt.Errorf("missing state cookie: %w", err)
	}

	var state statePayload
	if err := h.codec.Decode(stateCookieName, cookie.Value, &state); err != nil {
		return fmt.Errorf("decoding state cookie: %w", err)
	}

	if state.State == "" || state.State != queryState || time.Now().After(time.Unix(state.ExpiresAt, 0)) {
		return errors.New("invalid state")
	}

	return nil
}

func (h *AuthHandler) clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/api/auth/google",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func newAuthConfigFromEnv() authConfig {
	return authConfig{
		GoogleClientID:      strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		GoogleClientSecret:  strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		GoogleRedirectURL:   envOrDefault("GOOGLE_REDIRECT_URL", defaultGoogleRedirectURL),
		FrontendURL:         strings.TrimSpace(os.Getenv("FRONTEND_URL")),
		SessionCookieName:   envOrDefault("SESSION_COOKIE_NAME", defaultSessionCookieName),
		SessionCookieSecure: envBoolOrDefault("SESSION_COOKIE_SECURE", false),
		SessionTTL:          envHoursOrDefault("SESSION_TTL_HOURS", defaultSessionTTL),
		SessionSecret:       strings.TrimSpace(os.Getenv("SESSION_SECRET")),
		AllowedEmails:       parseAllowedEmails(os.Getenv("ALLOWED_EMAILS")),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envHoursOrDefault(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	hours, err := strconv.Atoi(value)
	if err != nil || hours <= 0 {
		return fallback
	}
	return time.Duration(hours) * time.Hour
}

func parseAllowedEmails(value string) map[string]struct{} {
	emails := make(map[string]struct{})
	for _, email := range strings.Split(value, ",") {
		email = strings.ToLower(strings.TrimSpace(email))
		if email != "" {
			emails[email] = struct{}{}
		}
	}
	return emails
}

func randomToken(byteCount int) (string, error) {
	bytes := make([]byte, byteCount)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
