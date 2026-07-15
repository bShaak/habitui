package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type BaseColors struct {
	Text       string `json:"text,omitempty"`
	Subtext    string `json:"subtext,omitempty"`
	Muted      string `json:"muted,omitempty"`
	Surface    string `json:"surface,omitempty"`
	SurfaceAlt string `json:"surface_alt,omitempty"`
	Primary    string `json:"primary,omitempty"`

	Red    string `json:"red,omitempty"`
	Orange string `json:"orange,omitempty"`
	Yellow string `json:"yellow,omitempty"`
	Green  string `json:"green,omitempty"`
	Blue   string `json:"blue,omitempty"`
	Purple string `json:"purple,omitempty"`
	Pink   string `json:"pink,omitempty"`
}

type Theme struct {
	Name string     `json:"name"`
	Base BaseColors `json:"base"`
}

type Config struct {
	Theme string     `json:"theme,omitempty"`
	Base  BaseColors `json:"base,omitzero"`
}

func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "habitui.config"
	}
	return filepath.Join(homeDir, ".habitui", "habitui.config")
}

func GetDefaultConfigPath() string {
	return "habitui.config"
}

func GetDefaultConfig() Config {
	return Config{Theme: DefaultThemeName}
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		configPath = GetDefaultConfigPath()
		data, err = os.ReadFile(configPath)
		if err != nil {
			defaultConfig := GetDefaultConfig()
			return &defaultConfig, nil
		}
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		defaultConfig := GetDefaultConfig()
		return &defaultConfig, nil
	}
	migrateLegacyColorKeys(data, &config)

	return &config, nil
}

// migrateLegacyColorKeys maps pre-theme-rename palette keys onto the semantic slots
// when the new key was not already set.
func migrateLegacyColorKeys(data []byte, config *Config) {
	if config == nil {
		return
	}
	var raw struct {
		Base map[string]string `json:"base"`
	}
	if err := json.Unmarshal(data, &raw); err != nil || len(raw.Base) == 0 {
		return
	}
	type mapping struct {
		legacy string
		dst    *string
	}
	for _, m := range []mapping{
		{"lavender", &config.Base.Primary},
		{"mauve", &config.Base.Purple},
		{"peach", &config.Base.Orange},
		{"overlay", &config.Base.Muted},
		{"surface1", &config.Base.Surface},
		{"surface2", &config.Base.SurfaceAlt},
	} {
		if *m.dst != "" {
			continue
		}
		if v := raw.Base[m.legacy]; v != "" {
			*m.dst = v
		}
	}
}

// SaveConfig writes config to ~/.habitui/habitui.config, creating the directory if needed.
func SaveConfig(config *Config) error {
	if config == nil {
		return nil
	}
	configPath := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(data, '\n'), 0o644)
}

func overlayString(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

func overlayColors(dst *BaseColors, src BaseColors) {
	overlayString(&dst.Text, src.Text)
	overlayString(&dst.Subtext, src.Subtext)
	overlayString(&dst.Muted, src.Muted)
	overlayString(&dst.Surface, src.Surface)
	overlayString(&dst.SurfaceAlt, src.SurfaceAlt)
	overlayString(&dst.Primary, src.Primary)
	overlayString(&dst.Red, src.Red)
	overlayString(&dst.Orange, src.Orange)
	overlayString(&dst.Yellow, src.Yellow)
	overlayString(&dst.Green, src.Green)
	overlayString(&dst.Blue, src.Blue)
	overlayString(&dst.Purple, src.Purple)
	overlayString(&dst.Pink, src.Pink)
}

func normalizeThemeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func GetTheme(config *Config) Theme {
	name := DefaultThemeName
	if config != nil && config.Theme != "" {
		name = config.Theme
	}
	theme := LookupTheme(name)
	if config != nil {
		overlayColors(&theme.Base, config.Base)
	}
	return theme
}

// LookupTheme returns a built-in theme by name, falling back to the default.
func LookupTheme(name string) Theme {
	name = normalizeThemeName(name)
	for _, t := range AllThemes {
		if t.Name == name {
			return t
		}
	}
	return CatppuccinMocha
}

// NextThemeName returns the next theme id after current.
func NextThemeName(current string) string {
	current = normalizeThemeName(current)
	if len(AllThemes) == 0 {
		return DefaultThemeName
	}
	for i, t := range AllThemes {
		if t.Name == current {
			return AllThemes[(i+1)%len(AllThemes)].Name
		}
	}
	return AllThemes[0].Name
}
