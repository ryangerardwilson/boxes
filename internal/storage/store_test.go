package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ryangerardwilson/boxes/internal/core"
)

func TestDefaultPathsUseSettingsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("BOXES_CONFIG", "")
	t.Setenv("BOXES_SETTINGS", "")
	t.Setenv("BOXES_DATABASE", "")
	t.Setenv("BOXES_DATA_HOME", "")

	paths, err := DefaultPaths()
	if err != nil {
		t.Fatal(err)
	}
	if paths.SettingsPath != filepath.Join(dir, ".config", "boxes", "config.json") {
		t.Fatalf("settings path = %q", paths.SettingsPath)
	}
	if paths.ConfigPath != filepath.Join(dir, "Documents", "notes", "rituals.txt") {
		t.Fatalf("config path = %q", paths.ConfigPath)
	}
	if paths.DatabasePath != filepath.Join(dir, "Data", "boxes.db") {
		t.Fatalf("database path = %q", paths.DatabasePath)
	}

	store := New(paths)
	if _, err := store.EnsureConfig(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(paths.SettingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	if settings.BoxesPath != paths.ConfigPath {
		t.Fatalf("settings boxes path = %q", settings.BoxesPath)
	}
	if settings.DatabasePath != paths.DatabasePath {
		t.Fatalf("settings database path = %q", settings.DatabasePath)
	}
	if info, err := os.Stat(filepath.Dir(paths.DatabasePath)); err != nil || !info.IsDir() {
		t.Fatalf("expected database directory to exist: %v", err)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := New(Paths{
		ConfigPath:   filepath.Join(dir, "config", "boxes.txt"),
		DatabasePath: filepath.Join(dir, "Data", "boxes.db"),
		DataHome:     filepath.Join(dir, "data"),
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
	if err := store.SaveDay(state, config); err != nil {
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

func TestSaveDayPersistsSQLiteHistory(t *testing.T) {
	dir := t.TempDir()
	store := New(Paths{
		ConfigPath:   filepath.Join(dir, "config", "boxes.txt"),
		DatabasePath: filepath.Join(dir, "Data", "boxes.db"),
		DataHome:     filepath.Join(dir, "data"),
	})
	config, err := core.ParseConfig([]byte("Move\nWork\n  Email\n  Deep work\n"))
	if err != nil {
		t.Fatal(err)
	}

	state := core.NewDayState("2026-06-09").WithChecked("work/email", true, config)
	if err := store.SaveDay(state, config); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", store.Paths.DatabasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var checked int
	if err := db.QueryRow(
		`select checked from box_daily_status where date = ? and item_id = ?`,
		"2026-06-09",
		"work/email",
	).Scan(&checked); err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked = %d, want 1", checked)
	}

	var statusRows int
	if err := db.QueryRow(`select count(*) from box_daily_status`).Scan(&statusRows); err != nil {
		t.Fatal(err)
	}
	if statusRows != 3 {
		t.Fatalf("status rows = %d, want 3 leaf rows", statusRows)
	}

	var eventRows int
	if err := db.QueryRow(`select count(*) from box_events where checked = 1`).Scan(&eventRows); err != nil {
		t.Fatal(err)
	}
	if eventRows != 1 {
		t.Fatalf("event rows = %d, want 1", eventRows)
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
