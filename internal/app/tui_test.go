package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ryangerardwilson/boxes/internal/core"
	"github.com/ryangerardwilson/boxes/internal/storage"
)

func TestMoveDateLoadsEmptyAdjacentDayWithoutSaving(t *testing.T) {
	store, config := newTUITestStore(t, "Move\n")
	model := NewModel(store, store.Paths.ConfigPath, config, core.NewDayState("2026-06-09"))

	model.moveDate(-1)

	if model.err != nil {
		t.Fatal(model.err)
	}
	if model.state.Date != "2026-06-08" {
		t.Fatalf("date = %q, want 2026-06-08", model.state.Date)
	}
	if len(model.state.CheckedIDs) != 0 {
		t.Fatalf("checked ids = %v, want none", model.state.CheckedIDs)
	}
	if _, err := os.Stat(store.Paths.DayPath("2026-06-08")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("navigation should not create day file, stat err = %v", err)
	}
}

func TestMoveDateLoadsSavedWeek(t *testing.T) {
	store, config := newTUITestStore(t, "Move\nWork\n  Email\n  Deep work\n")
	saved := core.NewDayState("2026-06-02").WithChecked("work/email", true, config)
	if err := store.SaveDay(saved, config); err != nil {
		t.Fatal(err)
	}
	model := NewModel(store, store.Paths.ConfigPath, config, core.NewDayState("2026-06-09"))

	model.moveDate(-7)

	if model.err != nil {
		t.Fatal(model.err)
	}
	if model.state.Date != "2026-06-02" {
		t.Fatalf("date = %q, want 2026-06-02", model.state.Date)
	}
	if !model.state.IsChecked("work/email") {
		t.Fatal("expected saved week state to be loaded")
	}
}

func TestUpdateNavigatesDaysAndWeeksFromKeys(t *testing.T) {
	store, config := newTUITestStore(t, "Move\n")
	model := NewModel(store, store.Paths.ConfigPath, config, core.NewDayState("2026-06-09"))

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = next.(Model)
	if model.state.Date != "2026-06-08" {
		t.Fatalf("h date = %q, want 2026-06-08", model.state.Date)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model = next.(Model)
	if model.state.Date != "2026-06-09" {
		t.Fatalf("l date = %q, want 2026-06-09", model.state.Date)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	model = next.(Model)
	if model.state.Date != "2026-06-02" {
		t.Fatalf("ctrl+h date = %q, want 2026-06-02", model.state.Date)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	model = next.(Model)
	if model.state.Date != "2026-06-09" {
		t.Fatalf("ctrl+l date = %q, want 2026-06-09", model.state.Date)
	}
}

func newTUITestStore(t *testing.T, configText string) (storage.Store, core.Config) {
	t.Helper()

	dir := t.TempDir()
	store := storage.New(storage.Paths{
		ConfigPath:   filepath.Join(dir, "config", "boxes.txt"),
		DatabasePath: filepath.Join(dir, "Data", "boxes.db"),
		DataHome:     filepath.Join(dir, "data"),
	})
	config, err := core.ParseConfig([]byte(configText))
	if err != nil {
		t.Fatal(err)
	}
	return store, config
}
