package branding

import "strings"

const (
	BackgroundSourcePreset  = "preset"
	DefaultBackgroundPreset = "canyon-falls"
)

var backgroundAssets = map[string]string{
	"canyon-falls":    "resources/images/canyon-falls.jpg",
	"fields-road":     "resources/images/fields-road.jpg",
	"mountain-valley": "resources/images/mountain-valley.jpg",
	"storm-valley":    "resources/images/storm-valley.jpg",
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
	return Config{
		Background: Background{
			Source: BackgroundSource{
				Type:    BackgroundSourcePreset,
				Variant: DefaultBackgroundPreset,
			},
		},
	}
}

func (c Config) Normalized() Config {
	if strings.TrimSpace(c.Background.Source.Type) != BackgroundSourcePreset {
		c.Background.Source.Type = BackgroundSourcePreset
	}

	variant := strings.TrimSpace(c.Background.Source.Variant)
	if _, ok := backgroundAssets[variant]; !ok {
		variant = DefaultBackgroundPreset
	}
	c.Background.Source.Variant = variant

	return c
}

func (c Config) BackgroundAsset() string {
	variant := c.Normalized().Background.Source.Variant
	return backgroundAssets[variant]
}
