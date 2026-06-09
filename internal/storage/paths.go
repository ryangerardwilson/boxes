package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Paths struct {
	SettingsPath      string
	ConfigPath        string
	LegacyConfigPath  string
	LegacyOutlinePath string
	DatabasePath      string
	DataHome          string
}

type Settings struct {
	BoxesPath    string `json:"boxes_path"`
	DatabasePath string `json:"database_path"`
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	configDir := filepath.Join(configHome, "boxes")

	dataHome := os.Getenv("BOXES_DATA_HOME")
	if dataHome == "" {
		dataHome = os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		dataHome = filepath.Join(dataHome, "boxes")
	}

	configOverride := os.Getenv("BOXES_CONFIG")
	settingsPath := os.Getenv("BOXES_SETTINGS")
	if settingsPath == "" && configOverride == "" {
		settingsPath = filepath.Join(configDir, "config.json")
	}
	settingsPath = expandPath(settingsPath, home)

	settings, settingsExists, settingsIsLegacy, err := loadSettings(settingsPath, home)
	if err != nil {
		return Paths{}, err
	}
	defaults := DefaultSettings(home)
	if settings.BoxesPath == "" {
		settings.BoxesPath = defaults.BoxesPath
	}
	if settings.DatabasePath == "" {
		settings.DatabasePath = defaults.DatabasePath
	}

	configPath := configOverride
	if configPath == "" {
		configPath = settings.BoxesPath
	}
	databasePath := os.Getenv("BOXES_DATABASE")
	if databasePath == "" {
		databasePath = settings.DatabasePath
	}

	legacyConfigPath := ""
	if settingsIsLegacy || !settingsExists {
		legacyConfigPath = settingsPath
	}
	legacyOutlinePath := filepath.Join(configDir, "boxes.txt")
	if expandPath(configPath, home) == expandPath(legacyOutlinePath, home) {
		legacyOutlinePath = ""
	}
	if configOverride != "" {
		legacyConfigPath = ""
		legacyOutlinePath = ""
	}

	return Paths{
		SettingsPath:      settingsPath,
		ConfigPath:        expandPath(configPath, home),
		LegacyConfigPath:  expandPath(legacyConfigPath, home),
		LegacyOutlinePath: expandPath(legacyOutlinePath, home),
		DatabasePath:      expandPath(databasePath, home),
		DataHome:          expandPath(dataHome, home),
	}, nil
}

func DefaultSettings(home string) Settings {
	return Settings{
		BoxesPath:    filepath.Join(home, "Documents", "notes", "rituals.txt"),
		DatabasePath: filepath.Join(home, "Data", "boxes.db"),
	}
}

func (p Paths) DayPath(date string) string {
	return filepath.Join(p.DataHome, "days", date+".json")
}

func loadSettings(path string, home string) (Settings, bool, bool, error) {
	if path == "" {
		return Settings{}, false, false, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Settings{}, false, false, nil
		}
		return Settings{}, false, false, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Settings{}, true, false, err
	}
	if _, ok := raw["items"]; ok {
		return Settings{}, true, true, nil
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, true, false, err
	}
	settings.BoxesPath = expandPath(settings.BoxesPath, home)
	settings.DatabasePath = expandPath(settings.DatabasePath, home)
	return settings, true, false, nil
}

func expandPath(path string, home string) string {
	if path == "" {
		return ""
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path
}
