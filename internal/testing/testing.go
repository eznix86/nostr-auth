package testing

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/eznix86/nostr-auth/internal/app"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/server"
)

// NewTestApp creates a test server.App instance with default test configuration
func NewTestApp() (*server.App, error) {
	return app.New(config.Config{
		AppURL:       "http://localhost:3000",
		ChallengeTTL: 5 * time.Minute,
		AppSecret:    "test-secret-change-in-tests",
		SessionTTL:   24 * time.Hour,
		CookieDomain: "",
		CookieSecure: false,
	})
}

// NewTestRequest creates a new HTTP request for testing
func NewTestRequest(method, url string) *http.Request {
	return httptest.NewRequest(method, url, nil)
}

// NewTestRecorder returns a response recorder for testing
func NewTestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// RequireNoError fails the test if err is not nil
func RequireNoError(t interface {
	Errorf(format string, args ...interface{})
}, err error) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
