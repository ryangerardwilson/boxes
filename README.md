# boxes

`boxes` is a daily checklist TUI.

Define an ordered set of daily boxes in config, then run `boxes` each day and mark what is done.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/ryangerardwilson/boxes/main/install.sh | bash
boxes version
boxes
```

The first bare `boxes` launch creates an editable outline config at `~/.config/boxes/boxes.txt` when one does not exist.

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

boxes today show
boxes today reset
boxes today check <id>
boxes today uncheck <id>
```

## Config

`boxes config open` opens the config file, creating an editable source-of-truth file when needed.

Default paths:

- config: `~/.config/boxes/boxes.txt`
- daily state: `~/.local/share/boxes/days/YYYY-MM-DD.json`

Environment overrides:

- `BOXES_CONFIG`: exact config file path
- `BOXES_DATA_HOME`: data directory for daily state

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

Legacy `~/.config/boxes/config.json` files are read and migrated to `boxes.txt` when the text config does not exist.

## TUI keys

```text
j/k or arrows  move
space/enter    check or uncheck
e              edit config in vim
r              reset today
?              toggle help
q              quit
```

## Implementation shape

The Go UI starts the component-library direction without pretending the library is done:

- `internal/components/l1`: design primitives and theme
- `internal/components/l2`: reusable TUI patterns
- `internal/components/l3`: boxes-specific day screen
- `internal/core`: config and day-state rules
- `internal/storage`: XDG paths, text config migration, and day-state persistence

## Maintainer Release

```sh
./push_release_upgrade.sh
```

This pushes the current commit, creates the next patch tag, publishes the Linux x64 release artifact, and installs that release from GitHub.
