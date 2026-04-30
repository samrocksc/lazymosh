# lazymosh — AGENTS.md

## What it does

A Charm-powered TUI for launching mosh (with SSH fallback) to saved servers. No passwords — relies entirely on SSH keys. Config lives at `~/.config/lazymosh/servers.json` (XDG compliant).

## Stack

- **Go 1.24** with `github.com/charmbracelet/bubbletea` v1.2.1 + `lipgloss`
- **Config**: `~/.config/lazymosh/servers.json` (XDG_CONFIG_HOME or ~/.config)
- **Auth**: SSH keys only, no password storage

## File layout

```
lazymosh/
├── main.go          # tea.Program entry, RootModel
├── model.go         # thin stub (types moved to pkg/)
├── cli/
│   └── cli.go       # flag parsing: --help, --version, --file/-f, --verbosity/-V
├── log/
│   └── log.go       # timestamped stderr logger, levels: error/warn/info/debug
├── pkg/
│   └── types.go     # Screen, NavigateMsg, ReloadMsg, ConnectErrMsg
├── config/
│   └── config.go    # Server/Store structs, Load/Save to JSON, SetPath() override
├── screens/
│   ├── list.go      # server list + connect/delete
│   ├── add.go       # add server form
│   └── edit.go      # edit server form + delete confirm
├── style/
│   └── style.go     # lipgloss color palette + helpers
└── lazymosh         # compiled binary
```

## CLI

```
lazymosh -h
```

| Flag | Description |
|------|-------------|
| `-h`, `--help` | show help and exit |
| `-v`, `--version` | show version and exit |
| `-V`, `--verbosity LEVEL` | log level: `error`, `warn`, `info`, `debug` (default: `info`) |
| `-f`, `--file PATH` | use PATH as servers.json (default: `~/.config/lazymosh/servers.json`) |

`--file` supports both `-f path` and `--file=path` syntax. `~` in PATH is expanded to `$HOME`.

### Examples

```bash
lazymosh                          # normal run
lazymosh -V debug                 # verbose logging to stderr
lazymosh -f ~/.my-servers.json    # alternate config file
lazymosh --file=~/servers.json    # = syntax also works
```

## Logging

All output goes to stderr (stdout is reserved for the TUI). Timestamped, level-prefixed:

```
10:04:58 DEBUG using default config file: /home/sam/.config/lazymosh/servers.json
10:04:58 DEBUG verbosity: debug, config: /home/sam/.config/lazymosh/servers.json
10:04:58 INFO  loaded 3 servers from /home/sam/.config/lazymosh/servers.json
10:04:58 DEBUG connecting: mosh root@1.2.3.4
10:04:58 INFO  mosh started — handing off (pid 12345)
```

Key logged events:
- Startup (config path resolved)
- Server load/save/delete
- mosh attempt + ssh fallback
- Errors (config parse, file write, etc.)

## Config schema

```json
{
  "servers": [
    {
      "id": "1",
      "name": "hetzner-berlin",
      "host": "1.2.3.4",
      "port": 22,
      "user": "root",
      "locality": "eu-berlin"
    }
  ]
}
```

## Screen flow

- **List** (`screenList`): Shows all servers. Navigate with ↑/↓ or j/k. Enter connects. A adds. E edits. D deletes (no confirm in list — edit screen has Y/N confirm).
- **Add** (`screenAdd`): Tab/Shift-Tab or ↑/↓ cycles focus through fields (name, host, user, port, locality). Enter on last field saves.
- **Edit** (`screenEdit`): Same as Add but pre-filled. D triggers delete confirm.

## Bug: message routing in Add/Edit screens

Background commands (save, delete) return typed messages like `SaveSuccessMsg`. In bubbletea, `Update()` receives all messages — **not just the ones you explicitly handle**. If a message type has no `case` in `Update()`'s outer switch, it falls through silently.

**Critical rule**: `SaveSuccessMsg`, `SaveErrMsg`, and `ReloadMsg` MUST have their own `case` in `Update()`, NOT inside `handleKey()` (which is only called from `case tea.KeyMsg:`). Putting them in `handleKey()` means they are silently dropped when fired from background goroutines.

Correct pattern:
```go
func (m AddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case SaveSuccessMsg:
        return m, func() tea.Msg { return pkg.NavigateMsg{Screen: pkg.ScreenList} }
    case tea.KeyMsg:
        return m.handleKey(msg)
    }
    return m, nil
}
```

## Connection logic (list.go → `connect()`)

1. Try `mosh user@host` — if it starts, hand off and exit
2. If mosh fails immediately, fall back to `ssh user@host`
3. Both use `Setctty + Setsid` to take over the terminal fully
4. No password support — SSH key must be authorized on target

## Key bindings

| Key | List | Add | Edit |
|-----|------|-----|------|
| ↑/↓ or j/k | navigate | change field focus | change field focus |
| Enter | connect | next field / save | next field / save |
| Tab / Shift+Tab | — | change field focus | change field focus |
| Backspace | — | delete char | delete char |
| A | → Add screen | — | — |
| E | → Edit screen | — | — |
| D | — | — | delete confirm |
| Y/N | — | — | confirm/deny delete |
| Esc | quit | → List | → List |
| Ctrl+C | quit | quit | quit |

## Color palette

Nightshade violet theme (style/style.go):
- Background: `#0d0d0d`
- Surface: `#1a1a2e`
- Border: `#2d2d44`
- Primary/Accent: `#7c3aed` / `#a78bfa` (violet)
- Success: `#34d399`
- Danger: `#f87171`
- Warning: `#fbbf24`
- Muted: `#6b7280`

## Build

```bash
go build -o lazymosh .
# or
make
```

## Running

```bash
./lazymosh                        # normal
lazymosh -V debug                 # verbose
lazymosh -f ~/.servers.json       # alternate config
```

First run creates `~/.config/lazymosh/servers.json` as an empty `{"servers":[]}`.
