package core

import "testing"

func TestParseConfigRejectsDuplicateIDs(t *testing.T) {
	_, err := ParseConfig([]byte(`{"items":[{"id":"move","label":"Move"},{"id":"move","label":"Move again"}]}`))
	if err == nil {
		t.Fatal("expected duplicate id error")
	}
}

func TestParseConfigAcceptsStringItems(t *testing.T) {
	config, err := ParseConfig([]byte(`{"items":["aaa","Plan tomorrow","Read/notes"]}`))
	if err != nil {
		t.Fatal(err)
	}

	want := []Item{
		{ID: "aaa", Label: "aaa"},
		{ID: "plan-tomorrow", Label: "Plan tomorrow"},
		{ID: "read-notes", Label: "Read/notes"},
	}
	if len(config.Items) != len(want) {
		t.Fatalf("expected %d items, got %d", len(want), len(config.Items))
	}
	for index, item := range want {
		if config.Items[index].ID != item.ID || config.Items[index].Label != item.Label {
			t.Fatalf("item %d = %#v, want %#v", index, config.Items[index], item)
		}
	}
}

func TestParseConfigDerivesMissingObjectIDFromLabel(t *testing.T) {
	config, err := ParseConfig([]byte(`{"items":[{"label":"Plan tomorrow"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if config.Items[0].ID != "plan-tomorrow" {
		t.Fatalf("expected derived id, got %q", config.Items[0].ID)
	}
}

func TestParseOutlineConfigAcceptsNestedLists(t *testing.T) {
	config, err := ParseConfig([]byte("Move\nWork\n  Email\n  Deep work\nPlan tomorrow\n"))
	if err != nil {
		t.Fatal(err)
	}

	flat := config.Flatten()
	wantIDs := []string{"move", "work", "work/email", "work/deep-work", "plan-tomorrow"}
	if len(flat) != len(wantIDs) {
		t.Fatalf("expected %d flat items, got %d", len(wantIDs), len(flat))
	}
	for index, wantID := range wantIDs {
		if flat[index].ID != wantID {
			t.Fatalf("flat id %d = %q, want %q", index, flat[index].ID, wantID)
		}
	}
	if config.LeafCount() != 4 {
		t.Fatalf("expected 4 leaf boxes, got %d", config.LeafCount())
	}
}

func TestParentStatusRequiresAllLeafChildren(t *testing.T) {
	config, err := ParseConfig([]byte("Work\n  Email\n  Deep work\n"))
	if err != nil {
		t.Fatal(err)
	}
	state := NewDayState("2026-06-09")

	state = state.WithChecked("work/email", true, config)
	if got := state.StatusFor("work", config); got != StatusPartial {
		t.Fatalf("expected partial parent, got %v", got)
	}

	state = state.WithChecked("work/deep-work", true, config)
	if got := state.StatusFor("work", config); got != StatusChecked {
		t.Fatalf("expected checked parent, got %v", got)
	}
}

func TestParentToggleUpdatesLeafChildrenOnly(t *testing.T) {
	config, err := ParseConfig([]byte("Work\n  Email\n  Deep work\n"))
	if err != nil {
		t.Fatal(err)
	}
	state := NewDayState("2026-06-09")
	state = state.Toggle("work", config)

	want := []string{"work/email", "work/deep-work"}
	if len(state.CheckedIDs) != len(want) {
		t.Fatalf("checked ids = %#v, want %#v", state.CheckedIDs, want)
	}
	for index, wantID := range want {
		if state.CheckedIDs[index] != wantID {
			t.Fatalf("checked id %d = %q, want %q", index, state.CheckedIDs[index], wantID)
		}
	}
}

func TestDayStatePreservesConfigOrder(t *testing.T) {
	config := Config{Items: []Item{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B"},
		{ID: "c", Label: "C"},
	}}
	state := NewDayState("2026-06-09")
	state = state.WithChecked("c", true, config)
	state = state.WithChecked("a", true, config)

	if got := state.CheckedIDs; len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Fatalf("checked ids not in config order: %#v", got)
	}
}
