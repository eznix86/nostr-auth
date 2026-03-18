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
	"github.com/eznix86/nostr-auth/internal/config"
	controller "github.com/eznix86/nostr-auth/internal/controller"
	"github.com/eznix86/nostr-auth/internal/nostr"
	serverpkg "github.com/eznix86/nostr-auth/internal/server"
	"github.com/eznix86/nostr-auth/internal/session"
)

func TestForwardAuthUnauthorizedNonBrowser(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestForwardAuthUnauthorizedNonBrowserPost(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/check", nil)

	app.Router.ServeHTTP(recorder, req)

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

	app.Router.ServeHTTP(recorder, req)

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

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestHealthzReturnsServiceUnavailableWhenDraining(t *testing.T) {
	app := newTestApp(t)
	app.SetReady(false)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestCSRFEndpointReturnsToken(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/csrf", nil)

	app.Router.ServeHTTP(recorder, req)

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
	if !strings.Contains(setCookieHeader, controller.CSRFCookieName+"=") {
		t.Fatal("expected csrf cookie to be set")
	}
}

func TestChallengeEndpointReturnsToken(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/challenge", nil)

	app.Router.ServeHTTP(recorder, req)

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
	if !strings.Contains(setCookieHeader, session.ChallengeCookieName+"=") {
		t.Fatal("expected challenge cookie to be set")
	}
}

func TestHomeStoresIntendedURLAndExposesItInProps(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"apps": {
				"default": {
					"config": {"domain": "demo.local"}
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
	req := httptest.NewRequest(http.MethodGet, "/?redirect=http://demo.local/private", nil)
	req.Header.Set("X-Inertia", "true")

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var intendedCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == session.IntendedURLCookieName {
			intendedCookie = cookie
			break
		}
	}

	if intendedCookie == nil {
		t.Fatal("expected intended url cookie to be set")
	}

	var payload struct {
		Props map[string]any `json:"props"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if got := payload.Props["intendedUrl"]; got != "http://demo.local/private" {
		t.Fatalf("props.intendedUrl = %#v, want %q", got, "http://demo.local/private")
	}

	if got, err := app.Controller.Session.VerifyIntendedURL(intendedCookie.Value); err != nil {
		t.Fatalf("VerifyIntendedURL() error = %v", err)
	} else if got != "http://demo.local/private" {
		t.Fatalf("intended url = %q, want %q", got, "http://demo.local/private")
	}
}

func TestVerifyChallengeSetsSignedSessionAndClearsChallenge(t *testing.T) {
	secretKey := nostrlib.Generate()
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"apps": {
				"default": {
					"config": {"domain": "private"},
					"users": ["`+secretKey.Public().Hex()+`"]
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
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
		ConfigFile:   configPath,
	})

	challengeToken, signedChallenge, err := issueChallengeForTest(app, "/private", 5*time.Minute)
	if err != nil {
		t.Fatalf("issueChallengeForTest() error = %v", err)
	}

	evt := nostrlib.Event{
		CreatedAt: nostrlib.Now(),
		Kind:      22242,
		Tags:      nostrlib.Tags{{"challenge", challengeToken}, {"relay", "127.0.0.1:3000"}},
	}
	if err := evt.Sign(secretKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	csrfToken, err := issueCsrfForTest(app)
	if err != nil {
		t.Fatalf("issueCsrfForTest() error = %v", err)
	}

	body, err := json.Marshal(controller.VerifyChallengeRequest{Event: evt.String()})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify", bytes.NewReader(body))
	req.Host = "127.0.0.1:3000"
	req.AddCookie(&http.Cookie{Name: session.ChallengeCookieName, Value: signedChallenge})
	req.AddCookie(&http.Cookie{Name: controller.CSRFCookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	if got := recorder.Header().Get("Location"); got != "/private" {
		t.Fatalf("location = %q, want %q", got, "/private")
	}

	var sessionCookie *http.Cookie
	var clearedChallenge bool
	var clearedIntendedURL bool
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == session.CookieName {
			sessionCookie = cookie
		}
		if cookie.Name == session.ChallengeCookieName && cookie.MaxAge == -1 {
			clearedChallenge = true
		}
		if cookie.Name == session.IntendedURLCookieName && cookie.MaxAge == -1 {
			clearedIntendedURL = true
		}
	}

	if sessionCookie == nil {
		t.Fatal("expected signed session cookie")
	}

	if !clearedChallenge {
		t.Fatal("expected challenge cookie to be cleared")
	}
	if !clearedIntendedURL {
		t.Fatal("expected intended url cookie to be cleared")
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	verifyReq.AddCookie(sessionCookie)
	if got := app.Controller.AuthenticatedPubkey(verifyReq); got != secretKey.Public().Hex() {
		t.Fatalf("authenticated pubkey = %q, want %q", got, secretKey.Public().Hex())
	}
}

func TestHomeRedirectsAuthenticatedUsersToLogoutPage(t *testing.T) {
	secretKey := nostrlib.Generate()
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(authCookie(t, app, secretKey.Public().Hex()))

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/logout" {
		t.Fatalf("Location = %q, want %q", got, "/logout")
	}
}

func TestHomeRedirectsAuthenticatedUsersToStoredIntendedURL(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"apps": {
				"default": {
					"config": {"domain": "demo.local"}
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

	secretKey := nostrlib.Generate()
	app := newTestAppWithConfig(t, config.Config{
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
		ConfigFile:   configPath,
	})

	intendedToken, err := app.Controller.Session.SignIntendedURL("http://demo.local/private", 5*time.Minute)
	if err != nil {
		t.Fatalf("SignIntendedURL() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(authCookie(t, app, secretKey.Public().Hex()))
	req.AddCookie(&http.Cookie{Name: session.IntendedURLCookieName, Value: intendedToken})

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "http://demo.local/private" {
		t.Fatalf("Location = %q, want %q", got, "http://demo.local/private")
	}
}

func TestLogoutPageRedirectsGuestsHome(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)

	app.Router.ServeHTTP(recorder, req)

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

	csrfToken, err := issueCsrfForTest(app)
	if err != nil {
		t.Fatalf("issueCsrfForTest() error = %v", err)
	}
	req.AddCookie(&http.Cookie{Name: controller.CSRFCookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	setCookieHeader := strings.Join(recorder.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(setCookieHeader, session.CookieName+"=") {
		t.Fatal("expected session cookie to be cleared")
	}
	if !strings.Contains(setCookieHeader, session.ChallengeCookieName+"=") {
		t.Fatal("expected challenge cookie to be cleared")
	}
	if !strings.Contains(setCookieHeader, session.IntendedURLCookieName+"=") {
		t.Fatal("expected intended url cookie to be cleared")
	}
	if !strings.Contains(setCookieHeader, controller.CSRFCookieName+"=") {
		t.Fatal("expected csrf cookie to be cleared")
	}
}

func TestLogoutPreservesExplicitRedirectForNextLogin(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"apps": {
				"default": {
					"config": {"domain": "demo.local"}
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
	req := httptest.NewRequest(http.MethodPost, "/logout?redirect=http://demo.local/private", nil)
	req.AddCookie(authCookie(t, app, nostrlib.Generate().Public().Hex()))

	csrfToken, err := issueCsrfForTest(app)
	if err != nil {
		t.Fatalf("issueCsrfForTest() error = %v", err)
	}
	req.AddCookie(&http.Cookie{Name: controller.CSRFCookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want %q", got, "/")
	}

	var intendedCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == session.IntendedURLCookieName && cookie.MaxAge != -1 {
			intendedCookie = cookie
			break
		}
	}

	if intendedCookie == nil {
		t.Fatal("expected intended url cookie to be re-set")
	}

	if got, err := app.Controller.Session.VerifyIntendedURL(intendedCookie.Value); err != nil {
		t.Fatalf("VerifyIntendedURL() error = %v", err)
	} else if got != "http://demo.local/private" {
		t.Fatalf("intended url = %q, want %q", got, "http://demo.local/private")
	}
}

func TestVerifyChallengeRejectedRedirectSetsFlashAndFallsBackHome(t *testing.T) {
	secretKey := nostrlib.Generate()
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"auth": {
			"enabled": true,
			"apps": {
				"default": {
					"config": {"domain": "demo.local"},
					"users": ["`+secretKey.Public().Hex()+`"]
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
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret",
		SessionTTL:   24 * time.Hour,
		ConfigFile:   configPath,
	})

	challengeToken, signedChallenge, err := issueChallengeForTest(app, "http://evil.local/private", 5*time.Minute)
	if err != nil {
		t.Fatalf("issueChallengeForTest() error = %v", err)
	}

	evt := nostrlib.Event{
		CreatedAt: nostrlib.Now(),
		Kind:      22242,
		Tags:      nostrlib.Tags{{"challenge", challengeToken}, {"relay", "127.0.0.1:3000"}},
	}
	if err := evt.Sign(secretKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	csrfToken, err := issueCsrfForTest(app)
	if err != nil {
		t.Fatalf("issueCsrfForTest() error = %v", err)
	}

	body, err := json.Marshal(controller.VerifyChallengeRequest{Event: evt.String()})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify", bytes.NewReader(body))
	req.Host = "127.0.0.1:3000"
	req.AddCookie(&http.Cookie{Name: session.ChallengeCookieName, Value: signedChallenge})
	req.AddCookie(&http.Cookie{Name: controller.CSRFCookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/" {
		t.Fatalf("location = %q, want %q", got, "/")
	}

	var flashCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "nostr_auth_flash" {
			flashCookie = cookie
			break
		}
	}

	if flashCookie == nil {
		t.Fatal("expected flash cookie in response")
	}

	homeRecorder := httptest.NewRecorder()
	homeReq := httptest.NewRequest(http.MethodGet, "/", nil)
	homeReq.Header.Set("X-Inertia", "true")
	homeReq.AddCookie(flashCookie)
	homeReq.AddCookie(authCookie(t, app, secretKey.Public().Hex()))

	app.Router.ServeHTTP(homeRecorder, homeReq)

	if homeRecorder.Code != http.StatusConflict {
		t.Fatalf("home status = %d, want %d", homeRecorder.Code, http.StatusConflict)
	}
	if got := homeRecorder.Header().Get("X-Inertia-Location"); got != "/logout" {
		t.Fatalf("home location = %q, want %q", got, "/logout")
	}
}

func TestVerifyChallengeRejectsInvalidCSRF(t *testing.T) {
	app := newTestApp(t)

	_, signedChallenge, err := issueChallengeForTest(app, "/private", 5*time.Minute)
	if err != nil {
		t.Fatalf("issueChallengeForTest() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify", bytes.NewReader([]byte(`{"csrfToken":"bad"}`)))
	req.AddCookie(&http.Cookie{Name: session.ChallengeCookieName, Value: signedChallenge})
	req.AddCookie(&http.Cookie{Name: controller.CSRFCookieName, Value: "good"})
	req.Header.Set("X-CSRF-Token", "bad")

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	if got := recorder.Header().Get("Location"); got != "/" {
		t.Fatalf("location = %q, want %q", got, "/")
	}

	setCookieHeader := strings.Join(recorder.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(setCookieHeader, "nostr_auth_flash=") {
		t.Fatal("expected flash cookie to be set with auth error")
	}

	var flashCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "nostr_auth_flash" {
			flashCookie = cookie
			break
		}
	}

	if flashCookie == nil {
		t.Fatal("expected flash cookie in response")
	}

	homeRecorder := httptest.NewRecorder()
	homeReq := httptest.NewRequest(http.MethodGet, "/", nil)
	homeReq.Header.Set("X-Inertia", "true")
	homeReq.AddCookie(flashCookie)

	app.Router.ServeHTTP(homeRecorder, homeReq)

	if homeRecorder.Code != http.StatusOK {
		t.Fatalf("home status = %d, want %d", homeRecorder.Code, http.StatusOK)
	}

	var payload struct {
		Flash map[string]any `json:"flash"`
	}
	if err := json.NewDecoder(homeRecorder.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if got := payload.Flash["error"]; got != "Your session expired. Please try again." {
		t.Fatalf("flash.error = %#v, want %q", got, "Your session expired. Please try again.")
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

	app.Router.ServeHTTP(recorder, req)

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

	app.Router.ServeHTTP(recorder, req)

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
				"admins": ["`+allowedKey.Public().Hex()+`"]
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

	app.Router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("X-Forwarded-Groups"); got != "admins" {
		t.Fatalf("X-Forwarded-Groups = %q, want %q", got, "admins")
	}
	if got := recorder.Header().Get("X-Auth-Request-Groups"); got != "admins" {
		t.Fatalf("X-Auth-Request-Groups = %q, want %q", got, "admins")
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

func TestBuildAssetsSetImmutableCacheControl(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/build/assets/app-DuvHv2Mf.js", nil)

	app.Router.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Cache-Control"); got != serverpkg.BuildCacheControl() {
		t.Fatalf("Cache-Control = %q, want %q", got, serverpkg.BuildCacheControl())
	}
}

func TestBuildAssetsSupportGzip(t *testing.T) {
	app := newTestApp(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/build/assets/app-DuvHv2Mf.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	app.Router.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want %q", got, "gzip")
	}
	if got := recorder.Header().Get("Vary"); !strings.Contains(got, "Accept-Encoding") {
		t.Fatalf("Vary = %q, want to contain %q", got, "Accept-Encoding")
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

func issueChallengeForTest(a *serverpkg.App, intendedURL string, ttl time.Duration) (token, signed string, err error) {
	token = "test-challenge-token"
	signed, err = a.Controller.Session.SignChallenge(token, intendedURL, ttl)
	return
}

func issueCsrfForTest(a *serverpkg.App) (string, error) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/csrf", nil)
	a.Router.ServeHTTP(recorder, req)

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == controller.CSRFCookieName {
			return cookie.Value, nil
		}
	}

	return "", nil
}

func authCookie(t *testing.T, app *serverpkg.App, pubkey string) *http.Cookie {
	t.Helper()

	recorder := httptest.NewRecorder()
	if !app.Controller.SetAuth(recorder, pubkey) {
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
	app.Controller.Account.Set(recorder, &nostr.Profile{NIP05: nip05})

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == account.ProfileCookieName {
			return cookie
		}
	}

	t.Fatal("expected profile cookie")
	return nil
}
