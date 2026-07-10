package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Theme struct {
	Name string     `json:"name"`
	Base BaseColors `json:"base"`
}

type BaseColors struct {
	Rosewater string `json:"rosewater"`
	Flamingo  string `json:"flamingo"`
	Mauve     string `json:"mauve"`
	Red       string `json:"red"`
	Peach     string `json:"peach"`
	Yellow    string `json:"yellow"`
	Green     string `json:"green"`
	Teal      string `json:"teal"`
	Sky       string `json:"sky"`
	Sapphire  string `json:"sapphire"`
	Blue      string `json:"blue"`
	Lavender  string `json:"lavender"`
	Text      string `json:"text"`
	Subtext   string `json:"subtext"`
	Overlay   string `json:"overlay"`
	Surface1  string `json:"surface1"`
	Surface2  string `json:"surface2"`
	Pink      string `json:"pink"`
}

type Config struct {
	Base BaseColors `json:",omitempty"`
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
	return Config{}
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

	return &config, nil
}

func overlayString(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

func overlayColors(dst *BaseColors, src BaseColors) {
	overlayString(&dst.Rosewater, src.Rosewater)
	overlayString(&dst.Flamingo, src.Flamingo)
	overlayString(&dst.Mauve, src.Mauve)
	overlayString(&dst.Red, src.Red)
	overlayString(&dst.Peach, src.Peach)
	overlayString(&dst.Yellow, src.Yellow)
	overlayString(&dst.Green, src.Green)
	overlayString(&dst.Teal, src.Teal)
	overlayString(&dst.Sky, src.Sky)
	overlayString(&dst.Sapphire, src.Sapphire)
	overlayString(&dst.Blue, src.Blue)
	overlayString(&dst.Lavender, src.Lavender)
	overlayString(&dst.Text, src.Text)
	overlayString(&dst.Subtext, src.Subtext)
	overlayString(&dst.Overlay, src.Overlay)
	overlayString(&dst.Surface1, src.Surface1)
	overlayString(&dst.Surface2, src.Surface2)
	overlayString(&dst.Pink, src.Pink)
}

func GetTheme(config *Config) Theme {
	theme := CatppuccinMocha
	if config != nil {
		overlayColors(&theme.Base, config.Base)
	}
	return theme
}
