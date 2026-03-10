package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SessionCookieName = "nostr_auth_session"
const CSRFCookieName = "nostr_auth_csrf"

var ErrInvalidSession = errors.New("invalid session")

type SessionClaims struct {
	PubKey    string `json:"pubkey"`
	ExpiresAt int64  `json:"expires_at"`
}

type SessionSigner struct {
	secret []byte
	ttl    time.Duration
}

func NewSessionSigner(secret string, ttl time.Duration) *SessionSigner {
	return &SessionSigner{secret: []byte(secret), ttl: ttl}
}

func (s *SessionSigner) Sign(pubkey string) (string, error) {
	claims := SessionClaims{
		PubKey:    pubkey,
		ExpiresAt: time.Now().Add(s.ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	encodedSig := base64.RawURLEncoding.EncodeToString(s.signature(encodedPayload))

	return fmt.Sprintf("%s.%s", encodedPayload, encodedSig), nil
}

func (s *SessionSigner) Verify(token string) (*SessionClaims, error) {
	payload, signature, ok := splitToken(token)
	if !ok {
		return nil, ErrInvalidSession
	}

	expectedSignature := s.signature(payload)

	providedSignature, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return nil, ErrInvalidSession
	}

	if !hmac.Equal(providedSignature, expectedSignature) {
		return nil, ErrInvalidSession
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrInvalidSession
	}

	var claims SessionClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidSession
	}

	if claims.PubKey == "" || time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrInvalidSession
	}

	return &claims, nil
}

func splitToken(token string) (payload string, signature string, ok bool) {
	payload, signature, ok = strings.Cut(token, ".")
	if !ok || payload == "" || signature == "" || strings.Contains(signature, ".") {
		return "", "", false
	}

	return payload, signature, true
}

func (s *SessionSigner) signature(payload string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}
