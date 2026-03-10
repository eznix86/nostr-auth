package csrf

import (
	"crypto/subtle"
	"net/http"

	"github.com/eznix86/nostr-auth/internal/challenge"
	"github.com/eznix86/nostr-auth/internal/cookie"
)

const CookieName = "nostr_auth_csrf"

type Guard struct {
	jar *cookie.Jar
}

func NewGuard(jar *cookie.Jar) *Guard {
	return &Guard{jar: jar}
}

func (g *Guard) Ensure(w http.ResponseWriter, r *http.Request) (string, error) {
	if token := g.jar.Value(r, CookieName); token != "" {
		return token, nil
	}

	token, err := challenge.NewSessionID()
	if err != nil {
		return "", err
	}

	g.jar.Set(w, CookieName, token, 0)
	return token, nil
}

func (g *Guard) Valid(r *http.Request, token string) bool {
	if token == "" {
		return false
	}

	cookieValue := g.jar.Value(r, CookieName)
	if cookieValue == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookieValue), []byte(token)) == 1
}

func (g *Guard) Clear(w http.ResponseWriter) {
	g.jar.Clear(w, CookieName)
}
