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

func GetTheme(config *Config) Theme {
	theme := CatppuccinMocha

	if config.Base.Text != "" {
		theme.Base.Text = config.Base.Text
	}
	if config.Base.Subtext != "" {
		theme.Base.Subtext = config.Base.Subtext
	}
	if config.Base.Overlay != "" {
		theme.Base.Overlay = config.Base.Overlay
	}
	if config.Base.Surface1 != "" {
		theme.Base.Surface1 = config.Base.Surface1
	}
	if config.Base.Surface2 != "" {
		theme.Base.Surface2 = config.Base.Surface2
	}
	if config.Base.Rosewater != "" {
		theme.Base.Rosewater = config.Base.Rosewater
	}
	if config.Base.Flamingo != "" {
		theme.Base.Flamingo = config.Base.Flamingo
	}
	if config.Base.Mauve != "" {
		theme.Base.Mauve = config.Base.Mauve
	}
	if config.Base.Red != "" {
		theme.Base.Red = config.Base.Red
	}
	if config.Base.Peach != "" {
		theme.Base.Peach = config.Base.Peach
	}
	if config.Base.Yellow != "" {
		theme.Base.Yellow = config.Base.Yellow
	}
	if config.Base.Green != "" {
		theme.Base.Green = config.Base.Green
	}
	if config.Base.Teal != "" {
		theme.Base.Teal = config.Base.Teal
	}
	if config.Base.Sky != "" {
		theme.Base.Sky = config.Base.Sky
	}
	if config.Base.Sapphire != "" {
		theme.Base.Sapphire = config.Base.Sapphire
	}
	if config.Base.Blue != "" {
		theme.Base.Blue = config.Base.Blue
	}
	if config.Base.Lavender != "" {
		theme.Base.Lavender = config.Base.Lavender
	}
	if config.Base.Pink != "" {
		theme.Base.Pink = config.Base.Pink
	}

	return theme
}
