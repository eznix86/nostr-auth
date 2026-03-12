package branding

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	assets "github.com/eznix86/nostr-auth"
)

const (
	BackgroundSourcePreset  = "preset"
	DefaultBackgroundPreset = "canyon-falls"
)

var (
	backgroundManifestOnce sync.Once
	backgroundManifest     Manifest
	backgroundManifestErr  error
)

type Manifest struct {
	Default  string                     `json:"default"`
	Variants map[string]ManifestVariant `json:"variants"`
}

type ManifestVariant struct {
	File string `json:"file"`
}

type Config struct {
	Background Background `json:"background"`
}

type Background struct {
	Source BackgroundSource `json:"source"`
}

type BackgroundSource struct {
	Type    string `json:"type"`
	Variant string `json:"variant"`
}

func DefaultConfig() Config {
	manifest, _ := loadManifest()
	defaultVariant := manifest.Default
	if defaultVariant == "" {
		defaultVariant = DefaultBackgroundPreset
	}

	return Config{
		Background: Background{
			Source: BackgroundSource{
				Type:    BackgroundSourcePreset,
				Variant: defaultVariant,
			},
		},
	}
}

func (c Config) Normalized() Config {
	manifest, _ := loadManifest()
	defaultVariant := manifest.Default
	if defaultVariant == "" {
		defaultVariant = DefaultBackgroundPreset
	}

	if strings.TrimSpace(c.Background.Source.Type) != BackgroundSourcePreset {
		c.Background.Source.Type = BackgroundSourcePreset
	}

	variant := strings.TrimSpace(c.Background.Source.Variant)
	if _, ok := manifest.Variants[variant]; !ok {
		variant = defaultVariant
	}
	c.Background.Source.Variant = variant

	return c
}

func (c Config) BackgroundAsset() string {
	manifest, err := loadManifest()
	if err != nil {
		return ""
	}

	variant := c.Normalized().Background.Source.Variant
	background, ok := manifest.Variants[variant]
	if !ok || background.File == "" {
		background = manifest.Variants[manifest.Default]
	}

	if background.File == "" {
		return ""
	}

	return "resources/images/" + background.File
}

func loadManifest() (Manifest, error) {
	backgroundManifestOnce.Do(func() {
		data, err := assets.AssetsFS.ReadFile("resources/images/images.json")
		if err != nil {
			backgroundManifestErr = fmt.Errorf("read images manifest: %w", err)
			return
		}

		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			backgroundManifestErr = fmt.Errorf("parse images manifest: %w", err)
			return
		}

		if manifest.Default == "" {
			backgroundManifestErr = fmt.Errorf("images manifest default is missing")
			return
		}
		if manifest.Variants == nil {
			backgroundManifestErr = fmt.Errorf("images manifest variants are missing")
			return
		}
		if _, ok := manifest.Variants[manifest.Default]; !ok {
			backgroundManifestErr = fmt.Errorf("images manifest default variant %q is not defined", manifest.Default)
			return
		}

		backgroundManifest = manifest
	})

	if backgroundManifestErr != nil {
		return Manifest{}, backgroundManifestErr
	}

	return backgroundManifest, nil
}
