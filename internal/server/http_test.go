package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	nostrlib "fiatjaf.com/nostr"
	"github.com/eznix86/nostr-auth/internal/account"
	"github.com/eznix86/nostr-auth/internal/app"
	"github.com/eznix86/nostr-auth/internal/challenge"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/csrf"
	"github.com/eznix86/nostr-auth/internal/nostr"
	serverpkg "github.com/eznix86/nostr-auth/internal/server"
	"github.com/eznix86/nostr-auth/internal/session"
)

func TestForwardAuthUnauthorizedNonBrowser(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestForwardAuthUnauthorizedNonBrowserPost(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/check", nil)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestForwardAuthRedirectsBrowser(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("X-Forwarded-Host", "demo.local")
	req.Header.Set("X-Forwarded-Uri", "/private")

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	want := "http://localhost:3000/?redirect=" + url.QueryEscape("http://demo.local/private")
	if got := recorder.Header().Get("Location"); got != want {
		t.Fatalf("location = %q, want %q", got, want)
	}
}

func TestForwardAuthIngressSubrequestReturnsUnauthorized(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("X-Original-Method", http.MethodGet)
	req.Header.Set("X-Original-URI", "/")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "demo.local")

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestHealthzReturnsServiceUnavailableWhenDraining(t *testing.T) {
	app := newTestApp(t)
	app.SetReady(false)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestCSRFEndpointReturnsToken(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/csrf", nil)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var payload map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload["token"] == "" {
		t.Fatal("expected csrf token in payload")
	}

	setCookieHeader := strings.Join(recorder.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(setCookieHeader, csrf.CookieName+"=") {
		t.Fatal("expected csrf cookie to be set")
	}
}

func TestChallengeEndpointReturnsToken(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/challenge", nil)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var payload map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload["token"] == "" {
		t.Fatal("expected challenge token in payload")
	}
	if payload["relay"] == "" {
		t.Fatal("expected relay in payload")
	}

	setCookieHeader := strings.Join(recorder.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(setCookieHeader, challenge.SessionCookieName+"=") {
		t.Fatal("expected challenge session cookie to be set")
	}
}

func TestVerifyChallengeSetsSignedSessionAndClearsChallenge(t *testing.T) {
	secretKey := nostrlib.Generate()
	app := newTestApp(t)

	challengeValue, err := app.Handlers.Challenge.Issue("challenge-session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	evt := nostrlib.Event{
		CreatedAt: nostrlib.Now(),
		Kind:      22242,
		Tags:      nostrlib.Tags{{"challenge", challengeValue.Token}, {"relay", "127.0.0.1:3000"}},
	}
	if err := evt.Sign(secretKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	csrfToken, err := app.Handlers.CSRF.Ensure(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("EnsureCSRFToken() error = %v", err)
	}

	body, err := json.Marshal(serverpkg.VerifyChallengeRequest{Event: evt.String(), RedirectTo: "/private"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify", bytes.NewReader(body))
	req.Host = "127.0.0.1:3000"
	req.AddCookie(&http.Cookie{Name: challenge.SessionCookieName, Value: "challenge-session"})
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	if got := recorder.Header().Get("Location"); got != "/private" {
		t.Fatalf("location = %q, want %q", got, "/private")
	}

	var sessionCookie *http.Cookie
	var clearedChallenge bool
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == session.CookieName {
			sessionCookie = cookie
		}
		if cookie.Name == challenge.SessionCookieName && cookie.MaxAge == -1 {
			clearedChallenge = true
		}
	}

	if sessionCookie == nil {
		t.Fatal("expected signed session cookie")
	}

	if !clearedChallenge {
		t.Fatal("expected challenge cookie to be cleared")
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	verifyReq.AddCookie(sessionCookie)
	if got := app.Handlers.AuthenticatedPubkey(verifyReq); got != secretKey.Public().Hex() {
		t.Fatalf("authenticated pubkey = %q, want %q", got, secretKey.Public().Hex())
	}
}

func TestHomeRedirectsAuthenticatedUsersToLogoutPage(t *testing.T) {
	secretKey := nostrlib.Generate()
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(authCookie(t, app, secretKey.Public().Hex()))

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/logout" {
		t.Fatalf("Location = %q, want %q", got, "/logout")
	}
}

func TestLogoutPageRedirectsGuestsHome(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want %q", got, "/")
	}
}

func TestLogoutClearsSessionAndChallengeCookies(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	csrfToken, err := app.Handlers.CSRF.Ensure(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("EnsureCSRFToken() error = %v", err)
	}
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	setCookieHeader := strings.Join(recorder.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(setCookieHeader, session.CookieName+"=") {
		t.Fatal("expected session cookie to be cleared")
	}
	if !strings.Contains(setCookieHeader, challenge.SessionCookieName+"=") {
		t.Fatal("expected challenge cookie to be cleared")
	}
	if !strings.Contains(setCookieHeader, csrf.CookieName+"=") {
		t.Fatal("expected csrf cookie to be cleared")
	}
}

func TestVerifyChallengeRejectsInvalidCSRF(t *testing.T) {
	app := newTestApp(t)
	if _, err := app.Handlers.Challenge.Issue("challenge-session"); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify", bytes.NewReader([]byte(`{"csrfToken":"bad"}`)))
	req.AddCookie(&http.Cookie{Name: challenge.SessionCookieName, Value: "challenge-session"})
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "good"})
	req.Header.Set("X-CSRF-Token", "bad")

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	if got := recorder.Header().Get("Location"); got != "/?auth_error=invalid_csrf" {
		t.Fatalf("location = %q, want %q", got, "/?auth_error=invalid_csrf")
	}
}

func TestForwardAuthForbiddenWhenAuthorizerRejectsUser(t *testing.T) {
	allowedKey := nostrlib.Generate()
	rejectedKey := nostrlib.Generate()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"groups": {"admins": ["`+allowedKey.Public().Hex()+`"]},
			"apps": {
				"default": {
					"config": {"domain": "demo.local"},
					"users": ["group:admins"]
				}
			}
		},
		"branding": {
			"background": {
				"source": {"type": "preset", "variant": "canyon-falls"}
			}
		}
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	app := newTestAppWithConfig(t, config.Config{
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
		ConfigFile:   configPath,
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("X-Forwarded-Host", "demo.local")
	req.AddCookie(authCookie(t, app, rejectedKey.Public().Hex()))

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestForwardAuthForbiddenWhenConfigFileIsMissing(t *testing.T) {
	app := newTestAppWithConfig(t, config.Config{
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
		ConfigFile:   filepath.Join(t.TempDir(), "missing.json"),
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("X-Forwarded-Host", "demo.local")
	req.AddCookie(authCookie(t, app, nostrlib.Generate().Public().Hex()))

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestForwardAuthSetsGroupHeaders(t *testing.T) {
	allowedKey := nostrlib.Generate()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"groups": {
				"admins": ["`+allowedKey.Public().Hex()+`", "group:staff"],
				"staff": ["alice@example.com"]
			},
			"apps": {
				"default": {
					"config": {"domain": "demo.local"},
					"users": ["group:admins"]
				}
			}
		},
		"branding": {
			"background": {
				"source": {"type": "preset", "variant": "fields-road"}
			}
		}
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	app := newTestAppWithConfig(t, config.Config{
		AppURL:              "http://localhost:3000",
		ChallengeTTL:        5 * time.Minute,
		AppSecret:           "test-secret",
		SessionTTL:          24 * time.Hour,
		ConfigFile:          configPath,
		ProfileFetchTimeout: time.Second,
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("X-Forwarded-Host", "demo.local")
	req.AddCookie(authCookie(t, app, allowedKey.Public().Hex()))
	req.AddCookie(profileCookie(t, app, "alice@example.com"))

	app.Routes().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("X-Forwarded-Groups"); got != "admins,staff" {
		t.Fatalf("X-Forwarded-Groups = %q, want %q", got, "admins,staff")
	}
	if got := recorder.Header().Get("X-Auth-Request-Groups"); got != "admins,staff" {
		t.Fatalf("X-Auth-Request-Groups = %q, want %q", got, "admins,staff")
	}
}

func TestNewAppIgnoresMissingConfigFile(t *testing.T) {
	app := newTestAppWithConfig(t, config.Config{
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
		ConfigFile:   filepath.Join(t.TempDir(), "missing.json"),
	})

	if app == nil {
		t.Fatal("expected app to be created")
	}
}

func newTestApp(t *testing.T) *serverpkg.App {
	t.Helper()

	return newTestAppWithConfig(t, config.Config{
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
	})
}

func newTestAppWithConfig(t *testing.T, cfg config.Config) *serverpkg.App {
	t.Helper()

	serverApp, err := app.New(cfg)
	if err != nil {
		t.Fatalf("app.New() error = %v", err)
	}

	return serverApp
}

func authCookie(t *testing.T, app *serverpkg.App, pubkey string) *http.Cookie {
	t.Helper()

	recorder := httptest.NewRecorder()
	if !app.Handlers.SetAuth(recorder, pubkey) {
		t.Fatal("expected auth session cookie")
	}

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == session.CookieName {
			return cookie
		}
	}

	t.Fatal("expected auth session cookie")
	return nil
}

func profileCookie(t *testing.T, app *serverpkg.App, nip05 string) *http.Cookie {
	t.Helper()

	recorder := httptest.NewRecorder()
	app.Handlers.Account.Set(recorder, &nostr.Profile{NIP05: nip05})

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == account.ProfileCookieName {
			return cookie
		}
	}

	t.Fatal("expected profile cookie")
	return nil
}
