# boxes Product Engineer

## Product Contract

`boxes` is a local-first daily checklist TUI.

Core loop:

```text
ordered outline config -> check/uncheck today's leaf boxes -> per-day state file
```

The text config file is the source of truth for the ordered list. Daily state
stores only checked leaf item ids for one local calendar day.

Canonical config is a plain text outline:

```text
Move
Work
  Email
  Deep work
Plan tomorrow
```

Indent with two spaces per nesting level. A parent box is considered done if
and only if every lower leaf box under it is checked. Toggling a parent checks
or unchecks all descendant leaf boxes.

IDs are derived from each item's path, for example `Work > Email` becomes
`work/email`. Legacy JSON config may be read only as a migration input; it is
not the canonical editable config.

## CLI Contract

Canonical user-facing actions:

```text
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

Do not add dash-prefixed user-facing action flags. Keep help example-led and
declarative.

## Storage Contract

Config path:

- `$BOXES_CONFIG` when set
- otherwise `$XDG_CONFIG_HOME/boxes/boxes.txt`
- otherwise `~/.config/boxes/boxes.txt`

Legacy migration input:

- `$XDG_CONFIG_HOME/boxes/config.json`
- `~/.config/boxes/config.json`

Data path:

- `$BOXES_DATA_HOME/days/YYYY-MM-DD.json` when set
- otherwise `$XDG_DATA_HOME/boxes/days/YYYY-MM-DD.json`
- otherwise `~/.local/share/boxes/days/YYYY-MM-DD.json`

## Release Contract

Public install command:

```sh
curl -fsSL https://raw.githubusercontent.com/ryangerardwilson/boxes/main/install.sh | bash
```

`install.sh` installs the latest GitHub release by default. Release artifacts
are Linux x64 tarballs named `boxes-linux-x64.tar.gz`. The app's release path is
owned by `./push_release_upgrade.sh`.

## UI Contract

Use Bubble Tea for the event loop and Lip Gloss for styling. Keep the component
layers explicit:

- L1: primitives and theme only
- L2: reusable TUI patterns
- L3: boxes day screen and domain-specific composition

The first app is allowed to be small. Do not turn it into a generic framework
until repeated boxes-like apps create real pressure.

TUI key `e` opens the canonical config file in `vim` and reloads the outline
after vim exits.
