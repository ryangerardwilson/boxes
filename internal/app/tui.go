package app

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ryangerardwilson/boxes/internal/components/l1"
	"github.com/ryangerardwilson/boxes/internal/components/l3"
	"github.com/ryangerardwilson/boxes/internal/core"
	"github.com/ryangerardwilson/boxes/internal/storage"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	Edit   key.Binding
	Reset  key.Binding
	Help   key.Binding
	Quit   key.Binding
}

type editorFinishedMsg struct {
	err error
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j", "down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space", "check"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit config"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q", "quit"),
		),
	}
}

type Model struct {
	keys       keyMap
	theme      l1.Theme
	store      storage.Store
	configPath string
	config     core.Config
	state      core.DayState
	selected   int
	showHelp   bool
	message    string
	err        error
}

func NewModel(store storage.Store, configPath string, config core.Config, state core.DayState) Model {
	return Model{
		keys:       defaultKeyMap(),
		theme:      l1.DefaultTheme(),
		store:      store,
		configPath: configPath,
		config:     config,
		state:      state,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Up):
			m.move(-1)
		case key.Matches(msg, m.keys.Down):
			m.move(1)
		case key.Matches(msg, m.keys.Toggle):
			m.toggleSelected()
		case key.Matches(msg, m.keys.Edit):
			return m, m.openConfigEditor()
		case key.Matches(msg, m.keys.Reset):
			m.state = m.state.Reset()
			m.save("reset today")
		}
	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.message = ""
			return m, nil
		}
		m.reloadConfig()
	}
	return m, nil
}

func (m Model) View() string {
	errText := ""
	if m.err != nil {
		errText = m.err.Error()
	}
	return l3.RenderDay(m.theme, l3.DayView{
		Date:       m.state.Date,
		ConfigPath: m.configPath,
		Config:     m.config,
		State:      m.state,
		Selected:   m.selected,
		ShowHelp:   m.showHelp,
		Message:    m.message,
		Error:      errText,
	})
}

func (m *Model) move(delta int) {
	items := m.config.Flatten()
	if len(items) == 0 {
		return
	}
	next := m.selected + delta
	if next < 0 {
		next = len(items) - 1
	}
	if next >= len(items) {
		next = 0
	}
	m.selected = next
	m.message = ""
}

func (m *Model) toggleSelected() {
	items := m.config.Flatten()
	if len(items) == 0 {
		return
	}
	if m.selected >= len(items) {
		m.selected = len(items) - 1
	}
	item := items[m.selected]
	wasChecked := m.state.StatusFor(item.ID, m.config) == core.StatusChecked
	m.state = m.state.Toggle(item.ID, m.config)
	verb := "checked"
	if wasChecked {
		verb = "unchecked"
	}
	m.save(fmt.Sprintf("%s %s", verb, item.Label))
}

func (m *Model) save(message string) {
	if err := m.store.SaveDay(m.state); err != nil {
		m.err = err
		m.message = ""
		return
	}
	m.err = nil
	m.message = message
}

func (m Model) openConfigEditor() tea.Cmd {
	return tea.ExecProcess(exec.Command("vim", m.configPath), func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m *Model) reloadConfig() {
	config, exists, err := m.store.LoadConfig()
	if err != nil {
		m.err = err
		m.message = ""
		return
	}
	if !exists {
		config = core.DefaultConfig()
	}

	m.config = config
	items := m.config.Flatten()
	if len(items) == 0 {
		m.selected = 0
	} else if m.selected >= len(items) {
		m.selected = len(items) - 1
	}
	m.err = nil
	m.message = "reloaded config"
}

func Run(store storage.Store, configPath string, config core.Config, date string) error {
	state, err := store.LoadDay(date)
	if err != nil {
		return err
	}
	program := tea.NewProgram(NewModel(store, configPath, config, state), tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func Today() string {
	return core.Today(time.Now())
}
