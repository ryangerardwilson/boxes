package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerCheckAndShow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BOXES_CONFIG", filepath.Join(dir, "config.json"))
	t.Setenv("BOXES_DATA_HOME", filepath.Join(dir, "data"))

	var out bytes.Buffer
	var errOut bytes.Buffer
	runner := NewRunner(&out, &errOut)
	runner.Now = func() time.Time {
		return time.Date(2026, 6, 9, 10, 0, 0, 0, time.Local)
	}

	if code := runner.Run([]string{"config", "path"}); code != 0 {
		t.Fatalf("config path failed: %s", errOut.String())
	}
	if code := runner.Run([]string{"today", "show"}); code == 0 {
		t.Fatal("today show should fail before config exists")
	}

	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte("Move\nWork\n  Email\n  Deep work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	errOut.Reset()
	out.Reset()
	if code := runner.Run([]string{"today", "check", "move"}); code != 0 {
		t.Fatalf("today check failed: %s", errOut.String())
	}
	if !strings.Contains(out.String(), "checked Move") {
		t.Fatalf("unexpected check output: %q", out.String())
	}

	out.Reset()
	if code := runner.Run([]string{"today", "show"}); code != 0 {
		t.Fatalf("today show failed: %s", errOut.String())
	}
	if !strings.Contains(out.String(), "[x] move  Move") {
		t.Fatalf("unexpected show output: %q", out.String())
	}
	if !strings.Contains(out.String(), "[ ]   work/email  Email") {
		t.Fatalf("unexpected nested show output: %q", out.String())
	}

	out.Reset()
	if code := runner.Run([]string{"today", "check", "work"}); code != 0 {
		t.Fatalf("today check parent failed: %s", errOut.String())
	}
	out.Reset()
	if code := runner.Run([]string{"today", "show"}); code != 0 {
		t.Fatalf("today show after parent check failed: %s", errOut.String())
	}
	if !strings.Contains(out.String(), "[x] work  Work") {
		t.Fatalf("expected parent to be checked: %q", out.String())
	}
	if !strings.Contains(out.String(), "3/3") {
		t.Fatalf("expected leaf progress, got: %q", out.String())
	}
}

func TestLoadOrCreateConfigCreatesEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "boxes.txt")
	t.Setenv("BOXES_CONFIG", configPath)
	t.Setenv("BOXES_DATA_HOME", filepath.Join(dir, "data"))

	var out bytes.Buffer
	var errOut bytes.Buffer
	runner := NewRunner(&out, &errOut)

	_, config, ok := runner.loadOrCreateConfig()
	if !ok {
		t.Fatalf("loadOrCreateConfig failed: %s", errOut.String())
	}
	if len(config.Items) != 0 {
		t.Fatalf("expected empty first-run config, got %#v", config.Items)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
}
