# cview

`cview` is a terminal UI for browsing Claude Code session history for the current working directory.

It maps the current directory to Claude's project history folder under `~/.claude/projects`, loads each `.jsonl` session file, and shows:

- a searchable session list
- session metadata such as timestamps and branch
- a transcript preview for the selected session

## Run

```bash
go run ./cmd/cview
```

## Controls

- Type to filter sessions
- `j` / `k` or arrow keys to move
- `q` to quit

## Notes

This repo depends on:

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles`
- `github.com/charmbracelet/lipgloss`

If `go mod tidy` fails in a restricted environment, run it again with network access to populate `go.sum`.
