package account

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/eznix86/nostr-auth/internal/cookie"
	"github.com/eznix86/nostr-auth/internal/nostr"
)

const ProfileCookieName = "nostr_auth_profile"

type Cookie struct {
	jar *cookie.Jar
}

func NewCookie(jar *cookie.Jar) *Cookie {
	return &Cookie{jar: jar}
}

func (c *Cookie) Set(w http.ResponseWriter, profile *nostr.Profile) {
	if profile == nil {
		return
	}

	data, err := json.Marshal(profile)
	if err != nil {
		return
	}

	c.jar.Set(w, ProfileCookieName, url.QueryEscape(string(data)), 86400*7)
}

func (c *Cookie) Read(r *http.Request) *nostr.Profile {
	value := c.jar.Value(r, ProfileCookieName)
	if value == "" {
		return nil
	}

	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return nil
	}

	var profile nostr.Profile
	if err := json.Unmarshal([]byte(decoded), &profile); err != nil {
		return nil
	}

	return &profile
}

func (c *Cookie) Clear(w http.ResponseWriter) {
	c.jar.Clear(w, ProfileCookieName)
}
