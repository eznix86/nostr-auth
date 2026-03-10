package session

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

const CookieName = "nostr_auth_session"

var ErrInvalid = errors.New("invalid session")

type Claims struct {
	PubKey    string `json:"pubkey"`
	ExpiresAt int64  `json:"expires_at"`
}

type Signer struct {
	secret []byte
	ttl    time.Duration
}

func NewSigner(secret string, ttl time.Duration) *Signer {
	return &Signer{secret: []byte(secret), ttl: ttl}
}

func (s *Signer) Sign(pubkey string) (string, error) {
	claims := Claims{PubKey: pubkey, ExpiresAt: time.Now().Add(s.ttl).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	encodedSignature := base64.RawURLEncoding.EncodeToString(s.signature(encodedPayload))
	return fmt.Sprintf("%s.%s", encodedPayload, encodedSignature), nil
}

func (s *Signer) Verify(token string) (*Claims, error) {
	payload, signature, ok := splitToken(token)
	if !ok {
		return nil, ErrInvalid
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return nil, ErrInvalid
	}

	if !hmac.Equal(providedSignature, s.signature(payload)) {
		return nil, ErrInvalid
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrInvalid
	}

	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalid
	}

	if claims.PubKey == "" || time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrInvalid
	}

	return &claims, nil
}

func (s *Signer) signature(payload string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}

func splitToken(token string) (payload string, signature string, ok bool) {
	payload, signature, ok = strings.Cut(token, ".")
	if !ok || payload == "" || signature == "" || strings.Contains(signature, ".") {
		return "", "", false
	}

	return payload, signature, true
}
