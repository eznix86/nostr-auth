package authorization

import (
	"testing"

	nostrlib "fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

func TestCompileDisabledReturnsNil(t *testing.T) {
	authorizer, err := Compile(FileConfig{})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if authorizer != nil {
		t.Fatal("Compile() should return nil when auth is disabled")
	}
}

func TestAuthorizerAllowsPubKeyAndNIP05(t *testing.T) {
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
	})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if !authorizer.Allowed("localhost:5500", pubkey.Hex(), "") {
		t.Fatal("Allowed() should match normalized pubkey on host with port")
	}

	if !authorizer.Allowed("localhost", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "Alice@Example.com") {
		t.Fatal("Allowed() should match normalized nip05")
	}

	if authorizer.Allowed("other.local", pubkey.Hex(), "") {
		t.Fatal("Allowed() should reject unmatched domains")
	}
	if authorizer.Allowed("localhost", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "") {
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
			"staff":  {"alice@example.com"},
		},
		Apps: map[string]AppConfig{
			"default": {
				Config: AppMatchConfig{Domain: "localhost"},
				Users:  []string{"group:admins"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	groups := authorizer.Groups("localhost", pubkey.Hex(), "alice@example.com")
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
	})
	if err == nil {
		t.Fatal("Compile() should reject unknown groups")
	}
}

func TestPolicyRejectsWhenAuthorizerMissing(t *testing.T) {
	policy := &Policy{}

	if policy.Allowed("localhost:8081", "any-pubkey", "") {
		t.Fatal("Allowed() should return false when auth is disabled")
	}
}
