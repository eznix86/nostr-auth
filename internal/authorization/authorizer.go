package authorization

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	appauth "github.com/eznix86/nostr-auth/internal/auth"
)

const groupPrefix = "group:"

type FileConfig struct {
	Auth   AuthSettings         `json:"auth"`
	Groups map[string][]string  `json:"groups"`
	Apps   map[string]AppConfig `json:"apps"`
}

type AuthSettings struct {
	Enabled bool `json:"enabled"`
}

type AppConfig struct {
	Config AppMatchConfig `json:"config"`
	Users  []string       `json:"users"`
}

type AppMatchConfig struct {
	Domain  string   `json:"domain"`
	Domains []string `json:"domains"`
}

type Authorizer struct {
	apps map[string]CompiledApp
}

type CompiledApp struct {
	Name    string
	Domains map[string]struct{}
	PubKeys map[string]struct{}
	NIP05s  map[string]struct{}
}

type principalSet struct {
	PubKeys map[string]struct{}
	NIP05s  map[string]struct{}
}

func LoadFile(path string) (*Authorizer, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return Compile(cfg)
}

func Compile(cfg FileConfig) (*Authorizer, error) {
	if !cfg.Auth.Enabled {
		return nil, nil
	}

	compiled := &Authorizer{apps: make(map[string]CompiledApp)}

	for appName, appCfg := range cfg.Apps {
		app, err := compileApp(appName, appCfg, cfg.Groups)
		if err != nil {
			return nil, err
		}

		for domain := range app.Domains {
			compiled.apps[domain] = app
		}
	}

	return compiled, nil
}

func (a *Authorizer) Allowed(host, pubkey, nip05 string) bool {
	if a == nil {
		return true
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

func (a *Authorizer) appForHost(host string) (CompiledApp, bool) {
	normalizedHost := normalizeDomain(host)
	if app, ok := a.apps[normalizedHost]; ok {
		return app, true
	}

	hostname := stripPort(normalizedHost)
	app, ok := a.apps[hostname]
	return app, ok
}

func compileApp(appName string, appCfg AppConfig, groups map[string][]string) (CompiledApp, error) {
	domains := compileDomains(appCfg.Config)
	if len(domains) == 0 {
		return CompiledApp{}, fmt.Errorf("app %q has no domains configured", appName)
	}

	principals, err := expandPrincipals(appCfg.Users, groups, nil)
	if err != nil {
		return CompiledApp{}, fmt.Errorf("app %q: %w", appName, err)
	}

	return CompiledApp{
		Name:    appName,
		Domains: domains,
		PubKeys: principals.PubKeys,
		NIP05s:  principals.NIP05s,
	}, nil
}

func compileDomains(cfg AppMatchConfig) map[string]struct{} {
	domains := make(map[string]struct{})

	if cfg.Domain != "" {
		domains[normalizeDomain(cfg.Domain)] = struct{}{}
	}

	for _, domain := range cfg.Domains {
		if domain == "" {
			continue
		}

		domains[normalizeDomain(domain)] = struct{}{}
	}

	return domains
}

func expandPrincipals(entries []string, groups map[string][]string, chain map[string]bool) (principalSet, error) {
	principals := principalSet{
		PubKeys: make(map[string]struct{}),
		NIP05s:  make(map[string]struct{}),
	}

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if strings.HasPrefix(entry, groupPrefix) {
			groupName := strings.TrimPrefix(entry, groupPrefix)
			if chain[groupName] {
				return principalSet{}, fmt.Errorf("group cycle detected for %q", groupName)
			}

			members, ok := groups[groupName]
			if !ok {
				return principalSet{}, fmt.Errorf("unknown group %q", groupName)
			}

			nextChain := cloneChain(chain)
			nextChain[groupName] = true

			expanded, err := expandPrincipals(members, groups, nextChain)
			if err != nil {
				return principalSet{}, err
			}

			mergePrincipalSets(&principals, expanded)
			continue
		}

		if strings.Contains(entry, "@") {
			principals.NIP05s[normalizeNIP05(entry)] = struct{}{}
			continue
		}

		pubkey, err := appauth.ParseAllowedPubKey(entry)
		if err != nil || pubkey == nil {
			return principalSet{}, fmt.Errorf("invalid principal %q", entry)
		}

		principals.PubKeys[strings.ToLower(pubkey.Hex())] = struct{}{}
	}

	return principals, nil
}

func mergePrincipalSets(dst *principalSet, src principalSet) {
	for pubkey := range src.PubKeys {
		dst.PubKeys[pubkey] = struct{}{}
	}

	for nip05 := range src.NIP05s {
		dst.NIP05s[nip05] = struct{}{}
	}
}

func cloneChain(chain map[string]bool) map[string]bool {
	if chain == nil {
		return make(map[string]bool)
	}

	clone := make(map[string]bool, len(chain))
	for key, value := range chain {
		clone[key] = value
	}

	return clone
}

func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}

func normalizeNIP05(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func stripPort(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	parsedHost, _, err := net.SplitHostPort(host)
	if err == nil {
		return parsedHost
	}

	return host
}
