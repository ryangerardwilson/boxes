package storage

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ConfigPath       string
	LegacyConfigPath string
	DataHome         string
}

func DefaultPaths() (Paths, error) {
	configPath := os.Getenv("BOXES_CONFIG")
	legacyConfigPath := ""
	if configPath == "" {
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return Paths{}, err
			}
			configHome = filepath.Join(home, ".config")
		}
		configDir := filepath.Join(configHome, "boxes")
		configPath = filepath.Join(configDir, "boxes.txt")
		legacyConfigPath = filepath.Join(configDir, "config.json")
	}

	dataHome := os.Getenv("BOXES_DATA_HOME")
	if dataHome == "" {
		dataHome = os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return Paths{}, err
			}
			dataHome = filepath.Join(home, ".local", "share")
		}
		dataHome = filepath.Join(dataHome, "boxes")
	}

	return Paths{ConfigPath: configPath, LegacyConfigPath: legacyConfigPath, DataHome: dataHome}, nil
}

func (p Paths) DayPath(date string) string {
	return filepath.Join(p.DataHome, "days", date+".json")
}
