package appconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eznix86/nostr-auth/internal/branding"
)

func TestLoadFileReturnsDefaultsWhenMissing(t *testing.T) {
	cfg, err := LoadFile(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if got := cfg.Branding.Background.Source.Type; got != branding.BackgroundSourcePreset {
		t.Fatalf("background source type = %q, want %q", got, branding.BackgroundSourcePreset)
	}
	if got := cfg.Branding.Background.Source.Variant; got != branding.DefaultBackgroundPreset {
		t.Fatalf("background variant = %q, want %q", got, branding.DefaultBackgroundPreset)
	}
}

func TestLoadFileNormalizesInvalidBackgroundSource(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
		"branding": {
			"background": {
				"source": {
					"type": "file",
					"variant": "unknown"
				}
			}
		}
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if got := cfg.Branding.Background.Source.Type; got != branding.BackgroundSourcePreset {
		t.Fatalf("background source type = %q, want %q", got, branding.BackgroundSourcePreset)
	}
	if got := cfg.Branding.Background.Source.Variant; got != branding.DefaultBackgroundPreset {
		t.Fatalf("background variant = %q, want %q", got, branding.DefaultBackgroundPreset)
	}
}
