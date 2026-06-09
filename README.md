# boxes

`boxes` is a daily checklist TUI.

Define an ordered set of daily boxes in config, then run `boxes` each day and mark what is done.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/ryangerardwilson/boxes/main/install.sh | bash
boxes version
boxes
```

The first bare `boxes` launch creates:

- settings: `~/.config/boxes/config.json`
- boxes outline: `~/Documents/notes/rituals.txt`
- analytics database directory: `~/Data/`

## Run From Source

```sh
mise x go@1.26.4 -- go run ./cmd/boxes
mise x go@1.26.4 -- go run ./cmd/boxes help
```

## Install From Source

```sh
./install.sh from .
boxes version
```

The installer writes the runtime under `~/.boxes/bin/boxes` and publishes the user-facing launcher at `~/.local/bin/boxes`.

## Usage

```sh
boxes
boxes help
boxes version
boxes upgrade

boxes config path
boxes config open
boxes config settings path
boxes config settings open
boxes config database path

boxes today show
boxes today reset
boxes today check <id>
boxes today uncheck <id>
```

## Config

`boxes config open` opens the boxes outline, creating an editable source-of-truth file when needed.

`boxes config settings open` opens the JSON settings file that controls where the outline and SQLite database live.

Default paths:

- settings: `~/.config/boxes/config.json`
- boxes outline: `~/Documents/notes/rituals.txt`
- analytics database: `~/Data/boxes.db`
- daily state: `~/.local/share/boxes/days/YYYY-MM-DD.json`

Environment overrides:

- `BOXES_SETTINGS`: exact settings file path
- `BOXES_CONFIG`: exact boxes outline path
- `BOXES_DATABASE`: exact SQLite database path
- `BOXES_DATA_HOME`: data directory for daily state

Settings file:

```json
{
  "boxes_path": "/home/ryan/Documents/notes/rituals.txt",
  "database_path": "/home/ryan/Data/boxes.db"
}
```

Example config:

```text
Move
Read
Work
  Email
  Deep work
Plan tomorrow
```

Indent with two spaces for nested boxes. A parent box is done only when all of its lower boxes are done.

IDs are derived from each item's path. In the example above, the CLI id for `Email` is `work/email`. Change labels carefully; changing a label changes the derived id for that item and its children.

Legacy list-style `~/.config/boxes/config.json` and `~/.config/boxes/boxes.txt` files are read and migrated to the configured outline path when needed.

The SQLite database stores daily leaf-box status in `box_daily_status` and append-only changes in `box_events`.

## TUI keys

```text
h/l            previous/next day
ctrl+h/l       previous/next week
j/k or arrows  move
space/enter    check or uncheck
e              edit config in vim
r              reset viewed day
?              toggle help
q              quit
```

## Implementation shape

The Go UI starts the component-library direction without pretending the library is done:

- `internal/components/l1`: design primitives and theme
- `internal/components/l2`: reusable TUI patterns
- `internal/components/l3`: boxes-specific day screen
- `internal/core`: config and day-state rules
- `internal/storage`: settings, outline migration, SQLite history, and day-state persistence

## Maintainer Release

```sh
./push_release_upgrade.sh
```

This pushes the current commit, creates the next patch tag, publishes the Linux x64 release artifact, and installs that release from GitHub.
