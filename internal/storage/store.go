package storage

import (
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
	data, err := os.ReadFile(s.Paths.ConfigPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return core.Config{}, false, err
		}
		if s.Paths.LegacyConfigPath == "" {
			return core.Config{}, false, nil
		}
		legacyData, legacyErr := os.ReadFile(s.Paths.LegacyConfigPath)
		if legacyErr != nil {
			if errors.Is(legacyErr, os.ErrNotExist) {
				return core.Config{}, false, nil
			}
			return core.Config{}, false, legacyErr
		}
		config, parseErr := core.ParseConfig(legacyData)
		if parseErr != nil {
			return core.Config{}, true, parseErr
		}
		if writeErr := s.WriteConfig(config); writeErr != nil {
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
	if _, err := os.Stat(s.Paths.ConfigPath); err == nil {
		return s.Paths.ConfigPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if s.Paths.LegacyConfigPath != "" {
		if legacyData, err := os.ReadFile(s.Paths.LegacyConfigPath); err == nil {
			config, parseErr := core.ParseConfig(legacyData)
			if parseErr != nil {
				return "", parseErr
			}
			if writeErr := s.WriteConfig(config); writeErr != nil {
				return "", writeErr
			}
			return s.Paths.ConfigPath, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
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

func (s Store) SaveDay(state core.DayState) error {
	path := s.Paths.DayPath(state.Date)
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
