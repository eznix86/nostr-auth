package authorization

import (
	"io"
	"testing"

	nostrlib "fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/rs/zerolog"
)

func TestCompileDisabledReturnsNil(t *testing.T) {
	authorizer, err := Compile(FileConfig{}, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if authorizer != nil {
		t.Fatal("Compile() should return nil when auth is disabled")
	}
}

func TestAuthorizerAllowsPubKeyAndIgnoresNIP05Entries(t *testing.T) {
	secretKey := nostrlib.Generate()
	pubkey := secretKey.Public()

	authorizer, err := Compile(FileConfig{
		Auth: AuthSettings{Enabled: true},
		Groups: map[string][]string{
			"admins": {
				nip19.EncodeNpub(pubkey),
				"alice@example.com",
			},
		},
		Apps: map[string]AppConfig{
			"default": {
				Config: AppMatchConfig{Domain: "localhost"},
				Users:  []string{"group:admins"},
			},
		},
	}, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if !authorizer.Allowed("localhost:5500", pubkey.Hex()) {
		t.Fatal("Allowed() should match normalized pubkey on host with port")
	}

	if authorizer.Allowed("other.local", pubkey.Hex()) {
		t.Fatal("Allowed() should reject unmatched domains")
	}
	if authorizer.Allowed("localhost", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff") {
		t.Fatal("Allowed() should reject users without a matching principal")
	}
}

func TestAuthorizerGroupsReturnsMatchedGroups(t *testing.T) {
	secretKey := nostrlib.Generate()
	pubkey := secretKey.Public()

	authorizer, err := Compile(FileConfig{
		Auth: AuthSettings{Enabled: true},
		Groups: map[string][]string{
			"admins": {nip19.EncodeNpub(pubkey), "group:staff"},
			"staff":  {nip19.EncodeNpub(pubkey)},
		},
		Apps: map[string]AppConfig{
			"default": {
				Config: AppMatchConfig{Domain: "localhost"},
				Users:  []string{"group:admins"},
			},
		},
	}, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	groups := authorizer.Groups("localhost", pubkey.Hex())
	if len(groups) != 2 || groups[0] != "admins" || groups[1] != "staff" {
		t.Fatalf("Groups() = %v, want [admins staff]", groups)
	}
}

func TestCompileRejectsUnknownGroup(t *testing.T) {
	_, err := Compile(FileConfig{
		Auth: AuthSettings{Enabled: true},
		Apps: map[string]AppConfig{
			"default": {
				Config: AppMatchConfig{Domain: "localhost"},
				Users:  []string{"group:missing"},
			},
		},
	}, zerolog.New(io.Discard))
	if err == nil {
		t.Fatal("Compile() should reject unknown groups")
	}
}

func TestAuthorizerRejectsWhenMissing(t *testing.T) {
	var authorizer *Authorizer

	if authorizer.Allowed("localhost:8081", "any-pubkey") {
		t.Fatal("Allowed() should return false when auth is disabled")
	}
}

func TestAuthorizerAllowsRedirectForConfiguredDomain(t *testing.T) {
	authorizer, err := Compile(FileConfig{
		Auth: AuthSettings{Enabled: true},
		Apps: map[string]AppConfig{
			"default": {
				Config: AppMatchConfig{Domain: "demo.local"},
			},
		},
	}, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if !authorizer.AllowsRedirect("https://demo.local/private") {
		t.Fatal("AllowsRedirect() should allow configured domain")
	}
	if authorizer.AllowsRedirect("https://evil.local/private") {
		t.Fatal("AllowsRedirect() should reject unknown domain")
	}
}

func TestAuthorizerAllowsRedirectForConfiguredDomainWithPort(t *testing.T) {
	authorizer, err := Compile(FileConfig{
		Auth: AuthSettings{Enabled: true},
		Apps: map[string]AppConfig{
			"default": {
				Config: AppMatchConfig{Domains: []string{"localhost:3000"}},
			},
		},
	}, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if !authorizer.AllowsRedirect("http://localhost:3000/dashboard") {
		t.Fatal("AllowsRedirect() should allow configured domain with port")
	}
	if authorizer.AllowsRedirect("http://localhost:4000/dashboard") {
		t.Fatal("AllowsRedirect() should reject unknown domain with different port")
	}
}
