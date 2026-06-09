package l3

import (
	"fmt"
	"strings"

	"github.com/ryangerardwilson/boxes/internal/components/l1"
	"github.com/ryangerardwilson/boxes/internal/components/l2"
	"github.com/ryangerardwilson/boxes/internal/core"
)

type DayView struct {
	Date       string
	ConfigPath string
	Config     core.Config
	State      core.DayState
	Selected   int
	ShowHelp   bool
	Message    string
	Error      string
}

func RenderDay(theme l1.Theme, view DayView) string {
	items := view.Config.Flatten()
	if len(items) == 0 {
		lines := []string{
			theme.Title.Render("boxes"),
			"",
			theme.Muted.Render("No boxes configured."),
			theme.Muted.Render("Press e to edit config"),
			theme.Muted.Render(fmt.Sprintf("Config: %s", view.ConfigPath)),
			"",
			l2.RenderHelpBar(theme, view.ShowHelp),
		}
		return theme.Base.Render(strings.Join(lines, "\n"))
	}

	rows := make([]string, 0, len(items))
	checked := view.State.CheckedSet()
	for index, item := range items {
		rows = append(rows, l2.RenderChecklistRow(theme, l2.ChecklistRow{
			ID:          item.ID,
			Label:       item.Label,
			Depth:       item.Depth,
			Status:      view.Config.StatusFor(item.ID, checked),
			Selected:    index == view.Selected,
			HasChildren: item.HasChildren,
		}))
	}

	header := fmt.Sprintf("boxes  %s  %s", view.Date, l2.RenderProgress(theme, view.State.CompletedCount(view.Config), view.Config.LeafCount()))
	lines := []string{
		theme.Title.Render(header),
		"",
		l2.JoinRows(rows),
		"",
		l2.RenderHelpBar(theme, view.ShowHelp),
	}
	if view.Message != "" {
		lines = append(lines, theme.Muted.Render(view.Message))
	}
	if view.Error != "" {
		lines = append(lines, theme.Error.Render(view.Error))
	}
	return theme.Base.Render(strings.Join(lines, "\n"))
}
