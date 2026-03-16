package controller

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/eznix86/nostr-auth/internal/session"
)

const CSRFCookieName = "nostr_auth_csrf"

func (h *Handler) AuthenticatedPubkey(r *http.Request) string {
	claims, ok := h.Claims(r)
	if !ok {
		return ""
	}

	return claims.PubKey
}

func (h *Handler) Claims(r *http.Request) (*session.Claims, bool) {
	token := h.Cookie.Value(r, session.CookieName)
	if token == "" {
		return nil, false
	}

	claims, err := h.Session.Verify(token)
	if err != nil {
		return nil, false
	}

	return claims, true
}

func (h *Handler) SetAuth(w http.ResponseWriter, pubkey string) bool {
	token, err := h.Session.Sign(pubkey)
	if err != nil {
		return false
	}

	h.Cookie.Set(w, session.CookieName, token, 0)
	return true
}

func (h *Handler) ClearAuth(w http.ResponseWriter) {
	h.Cookie.Clear(w, session.CookieName)
}

func (h *Handler) CompleteAuth(w http.ResponseWriter, pubkey string) bool {
	if !h.SetAuth(w, pubkey) {
		return false
	}

	h.Cookie.Clear(w, session.ChallengeCookieName)
	h.ClearIntendedURL(w)
	return true
}

func (h *Handler) Logout(w http.ResponseWriter) {
	h.ClearAuth(w)
	h.Cookie.Clear(w, session.ChallengeCookieName)
	h.ClearIntendedURL(w)
	h.clearCSRF(w)
	h.Account.Clear(w)
}

func (h *Handler) RefreshAuth(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := h.Claims(r)
	if !ok {
		return false
	}

	return h.SetAuth(w, claims.PubKey)
}

func (h *Handler) SetIntendedURL(w http.ResponseWriter, intendedURL string) bool {
	if intendedURL == "" {
		h.Cookie.Clear(w, session.IntendedURLCookieName)
		return true
	}

	token, err := h.Session.SignIntendedURL(intendedURL, h.Config.ChallengeTTL)
	if err != nil {
		return false
	}

	h.Cookie.Set(w, session.IntendedURLCookieName, token, 0)
	return true
}

func (h *Handler) IntendedURL(r *http.Request) string {
	token := h.Cookie.Value(r, session.IntendedURLCookieName)
	if token == "" {
		return ""
	}

	intendedURL, err := h.Session.VerifyIntendedURL(token)
	if err != nil {
		return ""
	}

	return intendedURL
}

func (h *Handler) ClearIntendedURL(w http.ResponseWriter) {
	h.Cookie.Clear(w, session.IntendedURLCookieName)
}

func (h *Handler) IssueChallenge(w http.ResponseWriter, intendedURL string) (string, error) {
	token, err := randomHex(16)
	if err != nil {
		return "", err
	}

	signed, err := h.Session.SignChallenge(token, intendedURL, h.Config.ChallengeTTL)
	if err != nil {
		return "", err
	}

	h.Cookie.Set(w, session.ChallengeCookieName, signed, 0)
	return token, nil
}

func (h *Handler) ensureCSRF(w http.ResponseWriter, r *http.Request) (string, error) {
	if token := h.Cookie.Value(r, CSRFCookieName); token != "" {
		return token, nil
	}

	token, err := randomHex(16)
	if err != nil {
		return "", err
	}

	h.Cookie.Set(w, CSRFCookieName, token, 0)
	return token, nil
}

func (h *Handler) validCSRF(r *http.Request, token string) bool {
	if token == "" {
		return false
	}

	cookieValue := h.Cookie.Value(r, CSRFCookieName)
	if cookieValue == "" {
		return false
	}

	return subtleConstantTimeCompare(cookieValue, token)
}

func (h *Handler) clearCSRF(w http.ResponseWriter) {
	h.Cookie.Clear(w, CSRFCookieName)
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
