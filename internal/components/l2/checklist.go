package l2

import (
	"fmt"
	"strings"

	"github.com/ryangerardwilson/boxes/internal/components/l1"
	"github.com/ryangerardwilson/boxes/internal/core"
)

type ChecklistRow struct {
	ID          string
	Label       string
	Depth       int
	Status      core.CompletionStatus
	Selected    bool
	HasChildren bool
}

func RenderChecklistRow(theme l1.Theme, row ChecklistRow) string {
	cursor := " "
	if row.Selected {
		cursor = ">"
	}

	box := "[ ]"
	labelStyle := theme.Unchecked
	if row.Status == core.StatusChecked {
		box = "[x]"
		labelStyle = theme.Checked
	} else if row.Status == core.StatusPartial {
		box = "[-]"
		labelStyle = theme.Progress
	}

	indent := strings.Repeat("  ", row.Depth)
	line := fmt.Sprintf("%s %s %s%s", cursor, box, indent, labelStyle.Render(row.Label))
	if row.Selected {
		return theme.Selected.Render(line)
	}
	return line
}

func RenderProgress(theme l1.Theme, done int, total int) string {
	if total == 0 {
		return theme.Muted.Render("no boxes configured")
	}
	return theme.Progress.Render(fmt.Sprintf("%d/%d", done, total))
}

func RenderHelpBar(theme l1.Theme, expanded bool) string {
	if expanded {
		return theme.Help.Render("j/k move  space check  e edit  r reset  ? hide help  q quit")
	}
	return theme.Help.Render("space check  e edit  ? help  q quit")
}

func JoinRows(rows []string) string {
	return strings.Join(rows, "\n")
}
