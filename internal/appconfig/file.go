package appconfig

import (
	"encoding/json"
	"os"

	"github.com/eznix86/nostr-auth/internal/authorization"
	"github.com/eznix86/nostr-auth/internal/branding"
)

type FileConfig struct {
	Auth     AuthConfig      `json:"auth"`
	Branding branding.Config `json:"branding"`
}

type AuthConfig struct {
	Enabled bool                               `json:"enabled"`
	Groups  map[string][]string                `json:"groups"`
	Apps    map[string]authorization.AppConfig `json:"apps"`
}

func Default() FileConfig {
	return FileConfig{
		Branding: branding.DefaultConfig(),
	}
}

func LoadFile(path string) (FileConfig, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return FileConfig{}, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, err
	}

	cfg.Branding = cfg.Branding.Normalized()
	return cfg, nil
}

func (a AuthConfig) PolicyConfig() authorization.FileConfig {
	return authorization.FileConfig{
		Auth:   authorization.AuthSettings{Enabled: a.Enabled},
		Groups: a.Groups,
		Apps:   a.Apps,
	}
}
