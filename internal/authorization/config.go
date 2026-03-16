package authorization

import (
	"encoding/json"
	"os"
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

func LoadFileConfig(path string) (FileConfig, error) {
	if path == "" {
		return FileConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return FileConfig{}, nil
		}

		return FileConfig{}, err
	}

	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, err
	}

	return cfg, nil
}
