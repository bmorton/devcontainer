// Package theme is the single source of truth for colors, glyphs and shared
// Lip Gloss styles used by every utility in this module. Utilities must not
// define their own hex colors; add to this package instead and document the
// addition in STYLEGUIDE.md.
package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Palette: neon cyberpunk on a near-black background. Keep in sync with
// STYLEGUIDE.md, which explains when to use each color.
var (
	Cyan    = lipgloss.Color("#00F0FF") // primary accent: titles, selection, keys
	Magenta = lipgloss.Color("#FF2BD6") // secondary accent: emphasis, badges, gradient end
	Violet  = lipgloss.Color("#9D4DFF") // tertiary accent: borders, dividers
	Acid    = lipgloss.Color("#39FF14") // success / ready
	Amber   = lipgloss.Color("#FFD400") // warning / unknown / attention
	Crimson = lipgloss.Color("#FF3864") // error / action required
	Snow    = lipgloss.Color("#EAF6FF") // primary text
	Steel   = lipgloss.Color("#9FA3C7") // secondary text
	Slate   = lipgloss.Color("#5C5F82") // muted text, hints, separators
	Night   = lipgloss.Color("#1A0B2E") // raised panel / selection background
	Ink     = lipgloss.Color("#0B0014") // deepest background, text on bright chips
)

// Glyphs: single-cell symbols with consistent meaning across utilities. Color
// never carries meaning on its own; a glyph (or word) always accompanies it.
const (
	GlyphReady    = "✓"
	GlyphNeeded   = "✗"
	GlyphOptional = "○"
	GlyphSkipped  = "–"
	GlyphUnknown  = "?"
	GlyphError    = "!"
	GlyphCursor   = "▌"
	GlyphChevron  = "⟫"
	GlyphBullet   = "◆"
	GlyphDot      = "·"
	GlyphArrow    = "→"
)

// Shared styles. Compose from these rather than calling lipgloss.NewStyle in
// feature code so the whole suite shifts together when the palette changes.
var (
	Text     = lipgloss.NewStyle().Foreground(Snow)
	Dim      = lipgloss.NewStyle().Foreground(Steel)
	Muted    = lipgloss.NewStyle().Foreground(Slate)
	Accent   = lipgloss.NewStyle().Foreground(Cyan)
	Emph     = lipgloss.NewStyle().Foreground(Magenta).Bold(true)
	Success  = lipgloss.NewStyle().Foreground(Acid).Bold(true)
	Warning  = lipgloss.NewStyle().Foreground(Amber).Bold(true)
	Danger   = lipgloss.NewStyle().Foreground(Crimson).Bold(true)
	Key      = lipgloss.NewStyle().Foreground(Ink).Background(Cyan).Bold(true).Padding(0, 1)
	KeyLabel = lipgloss.NewStyle().Foreground(Steel)
	Badge    = lipgloss.NewStyle().Foreground(Ink).Background(Magenta).Bold(true).Padding(0, 1)
	Chip     = lipgloss.NewStyle().Foreground(Steel).Background(Night).Padding(0, 1)
	Selected = lipgloss.NewStyle().Foreground(Snow).Background(Night).Bold(true)
	Panel    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Violet).Padding(0, 2)
	Code     = lipgloss.NewStyle().Foreground(Amber)
)

// Gradient colors s rune-by-rune across the given color stops (default
// Cyan→Magenta). Spaces are left uncolored so tracked-out titles stay crisp.
func Gradient(s string, stops ...color.Color) string {
	if len(stops) == 0 {
		stops = []color.Color{Cyan, Magenta}
	}
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	colors := lipgloss.Blend1D(len(runes), stops...)
	var b strings.Builder
	for i, r := range runes {
		if r == ' ' {
			b.WriteRune(r)
			continue
		}
		c := stops[len(stops)-1]
		if i < len(colors) {
			c = colors[i]
		}
		b.WriteString(lipgloss.NewStyle().Foreground(c).Bold(true).Render(string(r)))
	}
	return b.String()
}

// Tracked spaces letters out ("B O O T S T R A P") and widens word gaps so
// uppercase headings read as display type rather than shouting.
func Tracked(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		letters := strings.Split(w, "")
		words[i] = strings.Join(letters, " ")
	}
	return strings.Join(words, "   ")
}

// Title renders the canonical heading treatment: uppercase, tracked, gradient.
func Title(s string) string {
	return Gradient(Tracked(strings.ToUpper(s)))
}

// Header renders a utility's masthead: chevron + title, optional subtitle.
func Header(title, subtitle string) string {
	lines := []string{Accent.Render(GlyphChevron) + "  " + Title(title)}
	if subtitle != "" {
		lines = append(lines, Dim.Render(subtitle))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Divider renders a muted horizontal rule of the given width.
func Divider(width int) string {
	if width < 0 {
		width = 0
	}
	return Muted.Render(strings.Repeat("─", width))
}

// KeyHint renders a footer key binding, e.g. [enter] fix.
func KeyHint(key, label string) string {
	return Key.Render(key) + " " + KeyLabel.Render(label)
}

// Footer joins key hints with consistent spacing.
func Footer(hints ...string) string {
	return strings.Join(hints, "  ")
}

// StatusGlyph returns the colored glyph for a generic tri-state plus extras.
// Utilities with richer state machines should map onto these so the suite
// reads the same everywhere.
func StatusGlyph(kind string) string {
	switch kind {
	case "ready":
		return Success.Render(GlyphReady)
	case "needed":
		return Danger.Render(GlyphNeeded)
	case "optional":
		return Warning.Render(GlyphOptional)
	case "skipped":
		return Muted.Render(GlyphSkipped)
	case "error":
		return Danger.Render(GlyphError)
	default:
		return Warning.Render(GlyphUnknown)
	}
}
