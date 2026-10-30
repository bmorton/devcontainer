# Utilities style guide

Every TUI in `utilities/` should feel like one product: bright cyberpunk accents
on a near-black terminal, instant feedback, and a layout you can read at a
glance. This document is the contract; `internal/ui/theme` is its
implementation. If you need something that isn't here, add it to both.

## Palette

Dark terminals are the target. Never set a full-screen background color; let
the terminal's own black show through and paint only chips, badges and panels.

| Token     | Hex       | Use for                                                    |
| --------- | --------- | ---------------------------------------------------------- |
| `Cyan`    | `#00F0FF` | Primary accent: titles, selection cursor, key chips, links |
| `Magenta` | `#FF2BD6` | Secondary accent: emphasis, badges, gradient end           |
| `Violet`  | `#9D4DFF` | Borders, dividers, quiet structure                         |
| `Acid`    | `#39FF14` | Success, ready, done                                       |
| `Amber`   | `#FFD400` | Warning, unknown, optional-but-missing, notes              |
| `Crimson` | `#FF3864` | Error, action required                                     |
| `Snow`    | `#EAF6FF` | Primary text                                               |
| `Steel`   | `#9FA3C7` | Secondary text, summaries, descriptions                    |
| `Slate`   | `#5C5F82` | Muted hints, separators, placeholders                      |
| `Night`   | `#1A0B2E` | Raised backgrounds: selection row, blurred buttons, chips  |
| `Ink`     | `#0B0014` | Text on bright chips; the "background" we assume           |

Rules:

- **Color never carries meaning alone.** Pair every status color with a glyph
  or word (`✓ ready`, `✗ needed`). Lip Gloss degrades to 256/16 colors and
  `NO_COLOR` automatically; the UI must still make sense in monochrome.
- One gradient per screen: the title. `theme.Gradient` defaults to
  Cyan→Magenta; use Acid→Cyan only for a completion banner.
- No new hex values in feature code. Add a token to `theme.go` and a row here.

## Typography and glyphs

- Headings are uppercase, letter-spaced and gradient: `theme.Title("bootstrap")`
  renders `B O O T S T R A P`. Use `theme.Header(title, subtitle)` for a
  masthead; the subtitle is one sentence in `Dim`.
- Body text is sentence case. Avoid exclamation marks and emoji; use the glyph set:

  | Glyph | Meaning                       |
  | ----- | ----------------------------- |
  | `✓`   | ready / done (Acid)           |
  | `✗`   | required action (Crimson)     |
  | `○`   | optional action (Amber)       |
  | `–`   | skipped (Slate)               |
  | `?`   | unknown (Amber)               |
  | `!`   | error in the tool itself      |
  | `▌`   | selection cursor (Magenta)    |
  | `⟫`   | masthead chevron, "go" affordance |
  | `◆`   | panel heading bullet          |
  | `→`   | "pressing enter does…" hints  |

- Spinners: `spinner.MiniDot` in `theme.Accent`. Progress bars: Cyan→Magenta
  blend, no percentage text (the label beside it says `3/5 required ready`).

## Layout

- Page padding `1, 2`. Sections separated by one blank line in this order:
  masthead → progress/summary → primary list → details panel → footer → status.
- Lists: `cursor glyph title [optional] summary`. The selected row uses
  `theme.Selected` on the title only, never the whole line.
- Details live in `theme.Panel` (rounded Violet border, `0 2` padding). Panel
  heading is `◆ Title` in `Emph`.
- Footer is key chips: `theme.KeyHint("enter", "fix")`. Order: navigation,
  primary action, secondary actions, `q quit`. Show a chip only when its action
  is valid for the current selection.
- Clamp widths (`clamp(width-2, 40, 110)`); never assume 80 columns. Use
  `AltScreen` for interactive programs so quitting restores the shell cleanly.

## Behaviour

- **Checks run concurrently and never block the UI.** Each check gets a context
  with a timeout (30s) and reports a `Result` with a one-line summary.
- **Offer the fix next to the finding.** Anything that needs a real terminal
  (OAuth device flows, editors) runs through `tea.ExecProcess` and the UI
  re-checks afterwards. Mark fixes you cannot re-verify `Unverifiable` and say
  so in the summary.
- **Forms use huh with `theme.Huh()`.** Prefill from the current state and
  validate inline; `esc` cancels back to the list without side effects.
- Keys: `j/k ↑/↓` move, `enter`/`space` act, `r` refresh, `s` skip (optional
  only), `q`/`esc`/`ctrl+c` quit. Add bindings; don't change these.
- Every utility has a non-interactive mode (`--check` or similar) that prints
  the same information with the same glyphs and sets a meaningful exit code,
  and is used automatically when stdout is not a TTY.
- Status line (last row) is for transient feedback; clear it on the next key.

## Copy

- Summaries are lowercase fragments: `logged in to github.com as bmorton`,
  `missing user.email`. Details may be full sentences and may include commands
  in backticks.
- Say what will happen before it happens: `→ enter runs gh auth login`.
- Completion banner: `ALL SYSTEMS ONLINE`. Use it once, at the top.
