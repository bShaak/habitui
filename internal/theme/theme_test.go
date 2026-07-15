package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLookupTheme(t *testing.T) {
	got := LookupTheme("dracula")
	if got.Name != "dracula" {
		t.Fatalf("LookupTheme(dracula) name = %q", got.Name)
	}
	if got.Base.Primary == "" || got.Base.Text == "" {
		t.Fatal("LookupTheme(dracula) missing required colors")
	}

	fallback := LookupTheme("not-a-theme")
	if fallback.Name != DefaultThemeName {
		t.Fatalf("unknown theme fallback = %q, want %q", fallback.Name, DefaultThemeName)
	}
}

func TestNextThemeName(t *testing.T) {
	if len(AllThemes) < 2 {
		t.Fatal("expected multiple themes")
	}
	first := AllThemes[0].Name
	second := AllThemes[1].Name
	if NextThemeName(first) != second {
		t.Fatalf("NextThemeName(%q) = %q, want %q", first, NextThemeName(first), second)
	}
	last := AllThemes[len(AllThemes)-1].Name
	if NextThemeName(last) != first {
		t.Fatalf("NextThemeName wrap from %q = %q, want %q", last, NextThemeName(last), first)
	}
}

func TestGetThemeAppliesOverrides(t *testing.T) {
	cfg := &Config{
		Theme: "nord",
		Base: BaseColors{
			Primary: "#ffffff",
		},
	}
	got := GetTheme(cfg)
	if got.Name != "nord" {
		t.Fatalf("theme name = %q, want nord", got.Name)
	}
	if got.Base.Primary != "#ffffff" {
		t.Fatalf("primary override = %q, want #ffffff", got.Base.Primary)
	}
	if got.Base.Text == "" {
		t.Fatal("expected base text from nord")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	home := dir
	t.Setenv("HOME", home)

	cfg := &Config{Theme: "tokyo-night"}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	path := filepath.Join(home, ".habitui", "habitui.config")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected config at %s: %v", path, err)
	}
	if strings.Contains(string(data), `"base"`) {
		t.Fatalf("SaveConfig wrote empty base overrides: %s", data)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Theme != "tokyo-night" {
		t.Fatalf("loaded theme = %q, want tokyo-night", loaded.Theme)
	}
}

func TestSaveConfigOmitsEmptyBaseFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg := &Config{
		Theme: "nord",
		Base:  BaseColors{Primary: "#ffffff"},
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".habitui", "habitui.config"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, `"primary": "#ffffff"`) {
		t.Fatalf("missing primary override: %s", body)
	}
	if strings.Contains(body, `"text"`) {
		t.Fatalf("wrote empty text field: %s", body)
	}
}

func TestMigrateLegacyColorKeys(t *testing.T) {
	raw := []byte(`{
		"theme": "dracula",
		"base": {
			"lavender": "#aabbcc",
			"peach": "#ff9900",
			"mauve": "#cc88ff",
			"overlay": "#666666",
			"surface1": "#111111",
			"surface2": "#222222",
			"primary": "#0000ff"
		}
	}`)
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	migrateLegacyColorKeys(raw, &cfg)

	if cfg.Base.Primary != "#0000ff" {
		t.Fatalf("primary should keep new key, got %q", cfg.Base.Primary)
	}
	if cfg.Base.Orange != "#ff9900" {
		t.Fatalf("peach -> orange = %q", cfg.Base.Orange)
	}
	if cfg.Base.Purple != "#cc88ff" {
		t.Fatalf("mauve -> purple = %q", cfg.Base.Purple)
	}
	if cfg.Base.Muted != "#666666" {
		t.Fatalf("overlay -> muted = %q", cfg.Base.Muted)
	}
	if cfg.Base.Surface != "#111111" {
		t.Fatalf("surface1 -> surface = %q", cfg.Base.Surface)
	}
	if cfg.Base.SurfaceAlt != "#222222" {
		t.Fatalf("surface2 -> surface_alt = %q", cfg.Base.SurfaceAlt)
	}
}

func TestAllThemesHaveSemanticSlots(t *testing.T) {
	for _, th := range AllThemes {
		b := th.Base
		slots := map[string]string{
			"text": b.Text, "subtext": b.Subtext, "muted": b.Muted,
			"surface": b.Surface, "surface_alt": b.SurfaceAlt, "primary": b.Primary,
			"red": b.Red, "orange": b.Orange, "yellow": b.Yellow,
			"green": b.Green, "blue": b.Blue, "purple": b.Purple, "pink": b.Pink,
		}
		for name, val := range slots {
			if val == "" {
				t.Errorf("theme %s missing %s", th.Name, name)
			}
		}
	}
}
