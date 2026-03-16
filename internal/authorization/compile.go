package authorization

import (
	"fmt"
	"maps"
	"strings"

	appauth "github.com/eznix86/nostr-auth/internal/auth"
	"github.com/rs/zerolog"
)

type principalSet struct {
	PubKeys map[string]struct{}
}

type compiler struct {
	log zerolog.Logger
}

func Compile(cfg FileConfig, log zerolog.Logger) (*Authorizer, error) {
	if !cfg.Auth.Enabled {
		return nil, nil
	}

	return (&compiler{log: log}).compile(cfg)
}

func (c *compiler) compile(cfg FileConfig) (*Authorizer, error) {
	authorizer := &Authorizer{apps: make(map[string]AppPolicy)}

	for appName, appCfg := range cfg.Apps {
		appPolicy, err := c.compileAppPolicy(appName, appCfg, cfg.Groups)
		if err != nil {
			return nil, err
		}

		for domain := range appPolicy.Domains {
			authorizer.apps[domain] = appPolicy
		}
	}

	return authorizer, nil
}

func (c *compiler) compileAppPolicy(appName string, appCfg AppConfig, groups map[string][]string) (AppPolicy, error) {
	domains := compileDomains(appCfg.Config)
	if len(domains) == 0 {
		return AppPolicy{}, fmt.Errorf("app %q has no domains configured", appName)
	}

	principals, err := c.expandPrincipals(appCfg.Users, groups, nil)
	if err != nil {
		return AppPolicy{}, fmt.Errorf("app %q: %w", appName, err)
	}

	return AppPolicy{
		Name:    appName,
		Domains: domains,
		PubKeys: principals.PubKeys,
		Groups:  c.compileGroupsForApp(appCfg.Users, groups),
	}, nil
}

func (c *compiler) compileGroupsForApp(entries []string, groups map[string][]string) map[string]principalSet {
	compiled := make(map[string]principalSet)
	seen := make(map[string]bool)
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if !strings.HasPrefix(entry, groupPrefix) {
			continue
		}

		groupName := strings.TrimPrefix(entry, groupPrefix)
		c.collectCompiledGroups(groupName, groups, compiled, seen)
	}

	return compiled
}

func (c *compiler) collectCompiledGroups(groupName string, groups map[string][]string, compiled map[string]principalSet, seen map[string]bool) {
	if seen[groupName] {
		return
	}
	seen[groupName] = true

	members, ok := groups[groupName]
	if !ok {
		return
	}

	principals, err := c.expandPrincipals(members, groups, nil)
	if err == nil {
		compiled[groupName] = principals
	}

	for _, member := range members {
		member = strings.TrimSpace(member)
		if strings.HasPrefix(member, groupPrefix) {
			c.collectCompiledGroups(strings.TrimPrefix(member, groupPrefix), groups, compiled, seen)
		}
	}
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

func (c *compiler) expandPrincipals(entries []string, groups map[string][]string, chain map[string]bool) (principalSet, error) {
	principals := newPrincipalSet()

	for _, rawEntry := range entries {
		entry := strings.TrimSpace(rawEntry)
		if err := c.addPrincipalEntry(&principals, entry, groups, chain); err != nil {
			return principalSet{}, err
		}
	}

	return principals, nil
}

func (c *compiler) addPrincipalEntry(dst *principalSet, entry string, groups map[string][]string, chain map[string]bool) error {
	if entry == "" {
		return nil
	}

	if strings.HasPrefix(entry, groupPrefix) {
		expanded, err := c.expandGroup(strings.TrimPrefix(entry, groupPrefix), groups, chain)
		if err != nil {
			return err
		}

		mergePrincipalSets(dst, expanded)
		return nil
	}

	if strings.Contains(entry, "@") {
		c.log.Warn().Str("principal", entry).Msg("ignoring NIP-05 in authorization config")
		return nil
	}

	pubkey, err := appauth.ParseAllowedPubKey(entry)
	if err != nil || pubkey == nil {
		return fmt.Errorf("invalid principal %q", entry)
	}

	dst.PubKeys[strings.ToLower(pubkey.Hex())] = struct{}{}
	return nil
}

func (c *compiler) expandGroup(groupName string, groups map[string][]string, chain map[string]bool) (principalSet, error) {
	if chain[groupName] {
		return principalSet{}, fmt.Errorf("group cycle detected for %q", groupName)
	}

	members, ok := groups[groupName]
	if !ok {
		return principalSet{}, fmt.Errorf("unknown group %q", groupName)
	}

	nextChain := cloneChain(chain)
	nextChain[groupName] = true
	return c.expandPrincipals(members, groups, nextChain)
}

func newPrincipalSet() principalSet {
	return principalSet{
		PubKeys: make(map[string]struct{}),
	}
}

func mergePrincipalSets(dst *principalSet, src principalSet) {
	maps.Copy(dst.PubKeys, src.PubKeys)
}

func cloneChain(chain map[string]bool) map[string]bool {
	if chain == nil {
		return make(map[string]bool)
	}

	clone := make(map[string]bool, len(chain))
	maps.Copy(clone, chain)
	return clone
}
