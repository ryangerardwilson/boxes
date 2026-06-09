package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/ryangerardwilson/boxes/internal/core"
)

type Store struct {
	Paths Paths
}

func New(paths Paths) Store {
	return Store{Paths: paths}
}

func (s Store) LoadConfig() (core.Config, bool, error) {
	if err := s.EnsureSettings(); err != nil {
		return core.Config{}, false, err
	}

	data, err := os.ReadFile(s.Paths.ConfigPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return core.Config{}, false, err
		}
		legacyData, legacyErr := s.readLegacyConfig()
		if legacyErr != nil {
			return core.Config{}, false, legacyErr
		}
		if len(legacyData) == 0 {
			return core.Config{}, false, nil
		}
		config, parseErr := core.ParseConfig(legacyData)
		if parseErr != nil {
			return core.Config{}, true, parseErr
		}
		if writeErr := s.WriteConfig(config); writeErr != nil {
			return core.Config{}, true, writeErr
		}
		if writeErr := s.WriteSettings(); writeErr != nil {
			return core.Config{}, true, writeErr
		}
		return config, true, nil
	}
	config, err := core.ParseConfig(data)
	if err != nil {
		return core.Config{}, true, err
	}
	return config, true, nil
}

func (s Store) EnsureConfig() (string, error) {
	if err := s.EnsureSettings(); err != nil {
		return "", err
	}
	if _, err := os.Stat(s.Paths.ConfigPath); err == nil {
		return s.Paths.ConfigPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if legacyData, err := s.readLegacyConfig(); err != nil {
		return "", err
	} else if len(legacyData) > 0 {
		config, parseErr := core.ParseConfig(legacyData)
		if parseErr != nil {
			return "", parseErr
		}
		if writeErr := s.WriteConfig(config); writeErr != nil {
			return "", writeErr
		}
		if writeErr := s.WriteSettings(); writeErr != nil {
			return "", writeErr
		}
		return s.Paths.ConfigPath, nil
	}

	if err := os.MkdirAll(filepath.Dir(s.Paths.ConfigPath), 0o755); err != nil {
		return "", err
	}
	data := []byte(core.StarterConfigText())
	if err := os.WriteFile(s.Paths.ConfigPath, data, 0o644); err != nil {
		return "", err
	}
	return s.Paths.ConfigPath, nil
}

func (s Store) EnsureSettings() error {
	if s.Paths.SettingsPath == "" {
		return nil
	}
	if _, err := os.Stat(s.Paths.SettingsPath); err == nil {
		if s.Paths.LegacyConfigPath == s.Paths.SettingsPath {
			return nil
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return s.WriteSettings()
}

func (s Store) WriteSettings() error {
	if s.Paths.SettingsPath == "" {
		return nil
	}
	settings := Settings{
		BoxesPath:    s.Paths.ConfigPath,
		DatabasePath: s.Paths.DatabasePath,
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(s.Paths.SettingsPath), 0o755); err != nil {
		return err
	}
	if s.Paths.DatabasePath != "" {
		if err := os.MkdirAll(filepath.Dir(s.Paths.DatabasePath), 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(s.Paths.SettingsPath, data, 0o644)
}

func (s Store) WriteConfig(config core.Config) error {
	data, err := core.MarshalConfig(config)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Paths.ConfigPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.Paths.ConfigPath, data, 0o644)
}

func (s Store) readLegacyConfig() ([]byte, error) {
	for _, path := range []string{s.Paths.LegacyOutlinePath, s.Paths.LegacyConfigPath} {
		if path == "" || path == s.Paths.ConfigPath {
			continue
		}
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return nil, nil
}

func (s Store) LoadDay(date string) (core.DayState, error) {
	path := s.Paths.DayPath(date)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return core.NewDayState(date), nil
		}
		return core.DayState{}, err
	}
	return core.ParseDayState(data, date)
}

func (s Store) SaveDay(state core.DayState, config core.Config) error {
	path := s.Paths.DayPath(state.Date)
	previous, err := s.LoadDay(state.Date)
	if err != nil {
		return err
	}
	if err := s.SaveHistory(previous, state, config); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := core.MarshalDayState(state)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
