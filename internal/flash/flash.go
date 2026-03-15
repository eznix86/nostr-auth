package flash

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"

	"github.com/eznix86/nostr-auth/internal/cookie"
)

const (
	flashCookieName = "nostr_auth_flash"
	errorCookieName = "nostr_auth_errors"
)

type contextKey struct{}

type store struct {
	w   http.ResponseWriter
	r   *http.Request
	jar *cookie.Jar
}

func Middleware(jar *cookie.Jar) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := &store{w: w, r: r, jar: jar}
			ctx := context.WithValue(r.Context(), contextKey{}, s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type Provider struct{}

func (p *Provider) FlashErrors(ctx context.Context, errors gonertia.ValidationErrors) error {
	return writeCookie(ctx, errorCookieName, errors)
}

func (p *Provider) GetErrors(ctx context.Context) (gonertia.ValidationErrors, error) {
	return readCookie[gonertia.ValidationErrors](ctx, errorCookieName)
}

func (p *Provider) Flash(ctx context.Context, flash gonertia.Flash) error {
	return writeCookie(ctx, flashCookieName, flash)
}

func (p *Provider) GetFlash(ctx context.Context) (gonertia.Flash, error) {
	return readCookie[gonertia.Flash](ctx, flashCookieName)
}

func (p *Provider) ShouldClearHistory(ctx context.Context) (bool, error) {
	return false, nil
}

func (p *Provider) FlashClearHistory(ctx context.Context) error {
	return nil
}

func fromContext(ctx context.Context) *store {
	s, _ := ctx.Value(contextKey{}).(*store)
	return s
}

func writeCookie[T any](ctx context.Context, name string, payload T) error {
	s := fromContext(ctx)
	if s == nil {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	s.jar.Set(s.w, name, base64.URLEncoding.EncodeToString(data), 0)
	return nil
}

func readCookie[T any](ctx context.Context, name string) (T, error) {
	var zero T

	s := fromContext(ctx)
	if s == nil {
		return zero, nil
	}

	value := s.jar.Value(s.r, name)
	if value == "" {
		return zero, nil
	}

	s.jar.Clear(s.w, name)

	data, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return zero, nil
	}

	var payload T
	if err := json.Unmarshal(data, &payload); err != nil {
		return zero, nil
	}

	return payload, nil
}
