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
├── pkg/
│   └── types.go     # Screen, NavigateMsg, ReloadMsg, ConnectErrMsg
├── config/
│   └── servers.go   # Server/Store structs, Load/Save to JSON
├── screens/
│   ├── list.go      # server list + connect/delete
│   ├── add.go       # add server form
│   └── edit.go      # edit server form + delete confirm
├── style/
│   └── style.go     # lipgloss color palette + helpers
└── lazymosh         # compiled binary
```

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
./lazymosh
```

First run creates `~/.config/lazymosh/servers.json` as an empty `{"servers":[]}`.
