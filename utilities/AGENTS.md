# utilities/ — guidance for coding agents

Go module `github.com/bmorton/devcontainer/utilities`: small terminal tools that
ship inside the devcontainer image (built by `.devcontainer/Dockerfile`,
installed to `/usr/local/bin`). The first is `bootstrap`. Read
`STYLEGUIDE.md` before touching any UI.

## Layout and the pattern to copy

```
cmd/<tool>/main.go          thin: flags → internal/<tool>.Run(Options) → os.Exit
internal/<tool>/            all logic; exported Run(Options) int, RunCheck(...)
internal/<tool>/model.go    Bubble Tea model (Init/Update/View) only
internal/ui/theme/          shared palette, styles, glyphs, huh theme (the only place hex colors live)
internal/sh/                command runner interface + FakeRunner for tests
```

Adding a utility:

1. `cmd/<name>/main.go` with `--check`/`--version` flags, mirroring `cmd/bootstrap`.
2. `internal/<name>/` with `run.go` (TTY detection, program start), `model.go`,
   and the domain logic in plain functions that take `context.Context` and an
   injectable `Env`/runner so they are unit-testable without a shell.
3. Register the binary nowhere: the Dockerfile builds `./cmd/...` and installs
   every binary. Add it to the shim check list in `scripts/verify-devcontainer.sh`
   and to the README table.

## Libraries

- Bubble Tea **v2**: `charm.land/bubbletea/v2`. `View()` returns `tea.View`
  (`tea.NewView(s)`, set `AltScreen`, `WindowTitle`), keys arrive as
  `tea.KeyPressMsg` (match on `msg.String()` such as `"ctrl+c"`, `"enter"`),
  interactive subprocesses go through `tea.ExecProcess`.
- Lip Gloss v2 `charm.land/lipgloss/v2`, Bubbles v2 `charm.land/bubbles/v2/...`
  (spinner, progress), Huh v2 `charm.land/huh/v2` with `theme.Huh()`.
- Do not import `github.com/charmbracelet/...` v1 paths; they are a different API.
- Stdlib first. Avoid new dependencies unless they remove meaningful code.

## Conventions

- `gofmt`, `go vet`, tests green: `mise run lint && mise run test` (or the raw
  `go` commands). CI runs the same on every PR.
- Checks: concurrent, each with its own timeout, never writing to the terminal.
  Return a `Result{Status, Summary, Detail}`; summaries are short lowercase
  fragments, details may hold commands. Prefer parsing `--json` output over
  scraping text; when text is unavoidable, keep the regex next to the step and
  cover it with a test using `sh.FakeRunner`.
- Fixes: run the tool's own interactive command (`gh auth login`) rather than
  re-implementing a flow. Re-check afterwards; if that is impossible, set
  `Unverifiable` and say so.
- Never read or print secrets. Checking that a credential exists is fine;
  echoing tokens, keys or auth files is not.
- Keep `cmd/` free of logic so `internal/` stays testable and reusable.
- Exit codes: 0 ready, 1 required work outstanding, 2 usage error.

## Testing

- Unit tests for every step/check with `sh.FakeRunner` (see `steps_test.go`).
- Headless model tests drive `Update` with `tea.WindowSizeMsg`, result
  messages and `tea.KeyPressMsg{Code: 'j', Text: "j"}`, then assert on
  `View().Content` (see `model_test.go`). No real terminal is needed.
- For a visual check, run the binary inside a detached tmux session and
  `tmux capture-pane -p`; that is how screenshots in PRs are produced.

## Build and run

```sh
cd utilities
mise run build          # ./bin/<tool>
mise run test
mise run lint
go run ./cmd/bootstrap  # TUI
go run ./cmd/bootstrap --check
```

The image builds with the mise-managed Go version pinned in
`.devcontainer/mise.toml`; keep `go.mod`'s `go` directive at or below it.
