package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ryangerardwilson/boxes/internal/app"
	"github.com/ryangerardwilson/boxes/internal/core"
	"github.com/ryangerardwilson/boxes/internal/storage"
	"github.com/ryangerardwilson/boxes/internal/version"
)

type Runner struct {
	Out io.Writer
	Err io.Writer
	Now func() time.Time
}

func NewRunner(out io.Writer, errOut io.Writer) Runner {
	return Runner{
		Out: out,
		Err: errOut,
		Now: time.Now,
	}
}

func (r Runner) Run(args []string) int {
	if len(args) == 0 {
		return r.openTUI()
	}

	switch args[0] {
	case "help":
		if len(args) != 1 {
			return r.fail("help takes no arguments")
		}
		WriteHelp(r.Out)
		return 0
	case "version":
		if len(args) != 1 {
			return r.fail("version takes no arguments")
		}
		_, _ = fmt.Fprintln(r.Out, version.Version)
		return 0
	case "upgrade":
		if len(args) != 1 {
			return r.fail("upgrade takes no arguments")
		}
		return r.upgrade()
	case "config":
		return r.config(args[1:])
	case "today":
		return r.today(args[1:])
	default:
		return r.fail(fmt.Sprintf("unknown action %q\n\nRun: boxes help", args[0]))
	}
}

func (r Runner) openTUI() int {
	store, config, ok := r.loadOrCreateConfig()
	if !ok {
		return 1
	}
	if err := app.Run(store, store.Paths.ConfigPath, config, core.Today(r.Now())); err != nil {
		return r.fail(err.Error())
	}
	return 0
}

func (r Runner) config(args []string) int {
	if len(args) != 1 {
		return r.fail("config expects: path or open")
	}
	store, ok := r.store()
	if !ok {
		return 1
	}

	switch args[0] {
	case "path":
		_, _ = fmt.Fprintln(r.Out, store.Paths.ConfigPath)
		return 0
	case "open":
		path, err := store.EnsureConfig()
		if err != nil {
			return r.fail(err.Error())
		}
		if err := openEditor(path); err != nil {
			return r.fail(err.Error())
		}
		return 0
	default:
		return r.fail("config expects: path or open")
	}
}

func (r Runner) today(args []string) int {
	if len(args) == 0 {
		return r.fail("today expects: show, reset, check <id>, or uncheck <id>")
	}
	store, _, config, ok := r.loadConfig()
	if !ok {
		return 1
	}
	date := core.Today(r.Now())
	state, err := store.LoadDay(date)
	if err != nil {
		return r.fail(err.Error())
	}

	switch args[0] {
	case "show":
		if len(args) != 1 {
			return r.fail("today show takes no arguments")
		}
		return r.show(config, state)
	case "reset":
		if len(args) != 1 {
			return r.fail("today reset takes no arguments")
		}
		if err := store.SaveDay(state.Reset()); err != nil {
			return r.fail(err.Error())
		}
		_, _ = fmt.Fprintln(r.Out, "reset today")
		return 0
	case "check", "uncheck":
		if len(args) != 2 {
			return r.fail("today check/uncheck expects an item id")
		}
		id := args[1]
		item, exists := config.ItemByID(id)
		if !exists {
			return r.fail(fmt.Sprintf("unknown item id %q", id))
		}
		next := state.WithChecked(id, args[0] == "check", config)
		if err := store.SaveDay(next); err != nil {
			return r.fail(err.Error())
		}
		_, _ = fmt.Fprintf(r.Out, "%s %s\n", pastTense(args[0]), item.Label)
		return 0
	default:
		return r.fail("today expects: show, reset, check <id>, or uncheck <id>")
	}
}

func (r Runner) show(config core.Config, state core.DayState) int {
	checked := state.CheckedSet()
	for _, item := range config.Flatten() {
		mark := "[ ]"
		switch config.StatusFor(item.ID, checked) {
		case core.StatusChecked:
			mark = "[x]"
		case core.StatusPartial:
			mark = "[-]"
		}
		indent := strings.Repeat("  ", item.Depth)
		_, _ = fmt.Fprintf(r.Out, "%s %s%s  %s\n", mark, indent, item.ID, item.Label)
	}
	_, _ = fmt.Fprintf(r.Out, "%d/%d\n", state.CompletedCount(config), config.LeafCount())
	return 0
}

func (r Runner) upgrade() int {
	installer := os.Getenv("BOXES_INSTALLER")
	if installer != "" {
		cmd := exec.Command(installer, "upgrade")
		cmd.Stdout = r.Out
		cmd.Stderr = r.Err
		if err := cmd.Run(); err != nil {
			return r.fail(err.Error())
		}
		return 0
	}

	if _, err := os.Stat("./install.sh"); err == nil {
		installer = "./install.sh"
		cmd := exec.Command(installer, "upgrade")
		cmd.Stdout = r.Out
		cmd.Stderr = r.Err
		if err := cmd.Run(); err != nil {
			return r.fail(err.Error())
		}
		return 0
	}

	cmd := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/ryangerardwilson/boxes/main/install.sh | bash -s -- upgrade")
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	if err := cmd.Run(); err != nil {
		return r.fail(err.Error())
	}
	return 0
}

func (r Runner) loadConfig() (storage.Store, storage.Paths, core.Config, bool) {
	store, ok := r.store()
	if !ok {
		return storage.Store{}, storage.Paths{}, core.Config{}, false
	}
	config, exists, err := store.LoadConfig()
	if err != nil {
		_ = r.fail(err.Error())
		return storage.Store{}, storage.Paths{}, core.Config{}, false
	}
	if !exists {
		_, _ = fmt.Fprintf(r.Err, "No config found at %s\n", store.Paths.ConfigPath)
		_, _ = fmt.Fprintln(r.Err, "Run: boxes config open")
		return storage.Store{}, storage.Paths{}, core.Config{}, false
	}
	return store, store.Paths, config, true
}

func (r Runner) loadOrCreateConfig() (storage.Store, core.Config, bool) {
	store, ok := r.store()
	if !ok {
		return storage.Store{}, core.Config{}, false
	}
	config, exists, err := store.LoadConfig()
	if err != nil {
		_ = r.fail(err.Error())
		return storage.Store{}, core.Config{}, false
	}
	if exists {
		return store, config, true
	}
	if _, err := store.EnsureConfig(); err != nil {
		_ = r.fail(err.Error())
		return storage.Store{}, core.Config{}, false
	}
	return store, core.DefaultConfig(), true
}

func (r Runner) store() (storage.Store, bool) {
	paths, err := storage.DefaultPaths()
	if err != nil {
		_ = r.fail(err.Error())
		return storage.Store{}, false
	}
	return storage.New(paths), true
}

func (r Runner) fail(message string) int {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "unknown error"
	}
	_, _ = fmt.Fprintln(r.Err, message)
	return 1
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		return errors.New("set EDITOR or VISUAL to use boxes config open")
	}

	fields := strings.Fields(editor)
	cmd := exec.Command(fields[0], append(fields[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pastTense(verb string) string {
	if verb == "check" {
		return "checked"
	}
	return "unchecked"
}
