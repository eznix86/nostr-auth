package authorization

import (
	"net/url"
	"sort"
	"strings"
)

type Authorizer struct {
	apps map[string]AppPolicy
}

func (a *Authorizer) AllowsRedirect(target string) bool {
	if a == nil || target == "" {
		return false
	}

	parsedTarget, err := url.Parse(target)
	if err != nil {
		return false
	}
	if parsedTarget.Scheme != "http" && parsedTarget.Scheme != "https" {
		return false
	}

	host := parsedTarget.Hostname()
	if host == "" {
		return false
	}

	_, ok := a.appForHost(host)
	return ok
}

type AppPolicy struct {
	Name    string
	Domains map[string]struct{}
	PubKeys map[string]struct{}
	Groups  map[string]principalSet
}

func (a *Authorizer) Allowed(host, pubkey string) bool {
	if a == nil {
		return false
	}

	app, ok := a.appForHost(host)
	if !ok {
		return false
	}

	if _, ok := app.PubKeys[strings.ToLower(pubkey)]; ok {
		return true
	}

	if nip05 == "" {
		return false
	}

	_, ok = app.NIP05s[normalizeNIP05(nip05)]
	return ok
}

func (a *Authorizer) Groups(host, pubkey, nip05 string) []string {
	if a == nil {
		return nil
	}

	app, ok := a.appForHost(host)
	if !ok {
		return nil
	}

	normalizedPubkey := strings.ToLower(pubkey)
	normalizedNIP05 := normalizeNIP05(nip05)
	matched := make([]string, 0, len(app.Groups))
	for groupName, principals := range app.Groups {
		if _, ok := principals.PubKeys[normalizedPubkey]; ok {
			matched = append(matched, groupName)
			continue
		}
		if normalizedNIP05 != "" {
			if _, ok := principals.NIP05s[normalizedNIP05]; ok {
				matched = append(matched, groupName)
			}
		}
	}

	sort.Strings(matched)
	return matched
}

func (a *Authorizer) appForHost(host string) (AppPolicy, bool) {
	normalizedHost := normalizeDomain(host)
	if app, ok := a.apps[normalizedHost]; ok {
		return app, true
	}

	hostname := stripPort(normalizedHost)
	app, ok := a.apps[hostname]
	return app, ok
}
