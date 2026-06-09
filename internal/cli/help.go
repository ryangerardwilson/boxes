package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const helpText = `boxes

Daily checklist TUI from an editable outline config.

global actions:
  boxes help
    show this help
  boxes version
    print the installed version
  boxes upgrade
    upgrade when a release channel is configured

features:
  open today's checklist
  # boxes
  boxes

  edit or inspect the source-of-truth config
  # config path|open | config settings path|open | config database path
  boxes config path
  boxes config open
  boxes config settings open
  boxes config database path

  inspect or reset today's state
  # today show|reset
  boxes today show
  boxes today reset

  update today's state without opening the TUI
  # today check <id> | today uncheck <id>
  boxes today check move
  boxes today check work/email
  boxes today uncheck work
`

func WriteHelp(out io.Writer) {
	if file, ok := out.(*os.File); ok && term.IsTerminal(int(file.Fd())) && os.Getenv("NO_COLOR") == "" {
		_, _ = fmt.Fprintf(out, "\033[38;5;245m%s\033[0m", helpText)
		return
	}
	_, _ = io.WriteString(out, helpText)
}
