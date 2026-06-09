package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryangerardwilson/boxes/internal/core"
)

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := New(Paths{
		ConfigPath: filepath.Join(dir, "config", "config.json"),
		DataHome:   filepath.Join(dir, "data"),
	})

	if _, err := store.EnsureConfig(); err != nil {
		t.Fatal(err)
	}
	config, exists, err := store.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected config to exist")
	}
	config = core.Config{Items: []core.Item{{ID: "move", Label: "Move"}}}

	state := core.NewDayState("2026-06-09")
	state = state.WithChecked(config.Items[0].ID, true, config)
	if err := store.SaveDay(state); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.LoadDay("2026-06-09")
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.IsChecked(config.Items[0].ID) {
		t.Fatal("expected saved box to be checked")
	}
}

func TestLoadConfigMigratesLegacyJSONToTextConfig(t *testing.T) {
	dir := t.TempDir()
	store := New(Paths{
		ConfigPath:       filepath.Join(dir, "config", "boxes.txt"),
		LegacyConfigPath: filepath.Join(dir, "config", "config.json"),
		DataHome:         filepath.Join(dir, "data"),
	})
	if err := os.MkdirAll(filepath.Dir(store.Paths.LegacyConfigPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Paths.LegacyConfigPath, []byte(`{"items":["Move","Work"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	config, exists, err := store.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected legacy config to load")
	}
	if len(config.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(config.Items))
	}
	text, err := os.ReadFile(store.Paths.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != "Move\nWork\n" {
		t.Fatalf("unexpected migrated config: %q", string(text))
	}
}
