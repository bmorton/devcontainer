package bootstrap

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/bmorton/devcontainer/utilities/internal/ui/theme"
)

const checkTimeout = 30 * time.Second

type mode int

const (
	modeList mode = iota
	modeForm
)

type checkDoneMsg struct {
	idx int
	res Result
}

type fixDoneMsg struct {
	idx int
	err error
}

// Model is the Bubble Tea model for the interactive bootstrap.
type Model struct {
	env     Env
	steps   []*Step
	results []Result

	cursor        int
	width, height int
	spinner       spinner.Model
	progress      progress.Model

	mode       mode
	form       *huh.Form
	formIdx    int
	identName  string
	identEmail string

	status   string
	quitting bool
}

// NewModel builds the TUI model; all checks start running in Init.
func NewModel(env Env, steps []*Step) Model {
	results := make([]Result, len(steps))
	for i := range results {
		results[i] = Result{Status: StatusChecking}
	}
	return Model{
		env:      env,
		steps:    steps,
		results:  results,
		spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(theme.Accent)),
		progress: progress.New(progress.WithColors(theme.Cyan, theme.Magenta), progress.WithoutPercentage(), progress.WithWidth(40)),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkAll())
}

func (m Model) checkAll() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.steps))
	for i := range m.steps {
		cmds = append(cmds, m.check(i))
	}
	return tea.Batch(cmds...)
}

func (m Model) check(idx int) tea.Cmd {
	step, env := m.steps[idx], m.env
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
		defer cancel()
		return checkDoneMsg{idx: idx, res: step.Check(ctx, env)}
	}
}

// requiredProgress returns ready/total for required steps.
func (m Model) requiredProgress() (done, total int) {
	for i, s := range m.steps {
		if s.Optional {
			continue
		}
		total++
		if m.results[i].Status == StatusReady {
			done++
		}
	}
	return done, total
}

func (m Model) percent() float64 {
	done, total := m.requiredProgress()
	if total == 0 {
		return 1
	}
	return float64(done) / float64(total)
}

func (m Model) allRequiredReady() bool {
	done, total := m.requiredProgress()
	return done == total
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Messages handled regardless of mode.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.progress.SetWidth(clamp(m.width-6, 20, 60))
		if m.form != nil {
			m.form = m.form.WithWidth(clamp(m.width-8, 30, 80))
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm
		return m, cmd
	case checkDoneMsg:
		m.results[msg.idx] = msg.res
		return m, m.progress.SetPercent(m.percent())
	case fixDoneMsg:
		return m.afterFix(msg)
	}

	if m.mode == modeForm {
		return m.updateForm(msg)
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKey(key)
	}
	return m, nil
}

func (m Model) handleKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.status = ""
	switch key.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.steps)-1 {
			m.cursor++
		}
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = len(m.steps) - 1
	case "enter", " ":
		return m.runFix(m.cursor)
	case "r":
		for i := range m.results {
			m.results[i] = Result{Status: StatusChecking}
		}
		m.status = theme.Dim.Render("re-checking everything…")
		return m, m.checkAll()
	case "s":
		step := m.steps[m.cursor]
		if !step.Optional {
			m.status = theme.Warning.Render("required steps can't be skipped")
			return m, nil
		}
		if m.results[m.cursor].Status == StatusSkipped {
			m.results[m.cursor] = Result{Status: StatusChecking}
			return m, m.check(m.cursor)
		}
		m.results[m.cursor] = Result{Status: StatusSkipped, Summary: "skipped"}
	}
	return m, nil
}

func (m Model) runFix(idx int) (tea.Model, tea.Cmd) {
	step, res := m.steps[idx], m.results[idx]
	switch res.Status {
	case StatusChecking:
		m.status = theme.Dim.Render("still checking " + step.Title + "…")
		return m, nil
	case StatusReady:
		m.status = theme.Success.Render(step.Title + " is already ready")
		return m, nil
	}
	switch step.Fix.Kind {
	case FixCommand:
		m.status = theme.Dim.Render("running " + step.Fix.Label + "…")
		cmd := exec.Command(step.Fix.Command[0], step.Fix.Command[1:]...)
		return m, tea.ExecProcess(cmd, func(err error) tea.Msg { return fixDoneMsg{idx: idx, err: err} })
	case FixGitIdentity:
		return m.startIdentityForm(idx)
	default:
		m.status = theme.Warning.Render("no automatic fix; see the details panel")
		return m, nil
	}
}

func (m Model) afterFix(msg fixDoneMsg) (tea.Model, tea.Cmd) {
	step := m.steps[msg.idx]
	if msg.err != nil {
		m.status = theme.Danger.Render(fmt.Sprintf("%s exited with an error: %v", step.Fix.Label, msg.err))
		m.results[msg.idx] = Result{Status: StatusChecking}
		return m, m.check(msg.idx)
	}
	if step.Fix.Unverifiable {
		m.results[msg.idx] = Ready("completed (not verifiable from here)")
		m.status = theme.Success.Render(step.Title + " done")
		return m, m.progress.SetPercent(m.percent())
	}
	m.status = theme.Dim.Render("re-checking " + step.Title + "…")
	m.results[msg.idx] = Result{Status: StatusChecking}
	return m, m.check(msg.idx)
}

// --- git identity form -----------------------------------------------------------

func (m Model) startIdentityForm(idx int) (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m.identName, m.identEmail = identity(ctx, m.env)
	m.formIdx = idx
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Description("Shown as the author on your commits.").
				Placeholder("Ada Lovelace").
				Value(&m.identName).
				Validate(nonEmpty("name")),
			huh.NewInput().
				Title("Email").
				Description("Use the address attached to your GitHub account.").
				Placeholder("ada@example.com").
				Value(&m.identEmail).
				Validate(func(s string) error {
					if err := nonEmpty("email")(s); err != nil {
						return err
					}
					if !strings.Contains(s, "@") {
						return fmt.Errorf("that doesn't look like an email address")
					}
					return nil
				}),
		),
	).WithTheme(theme.Huh()).WithShowHelp(true).WithWidth(clamp(m.width-8, 30, 80))
	m.mode = modeForm
	return m, m.form.Init()
}

func nonEmpty(what string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", what)
		}
		return nil
	}
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	fm, cmd := m.form.Update(msg)
	if f, ok := fm.(*huh.Form); ok {
		m.form = f
	}
	switch m.form.State {
	case huh.StateCompleted:
		m.mode = modeList
		m.status = theme.Dim.Render("saving git identity…")
		m.results[m.formIdx] = Result{Status: StatusChecking}
		return m, m.applyIdentity(m.formIdx, m.identName, m.identEmail)
	case huh.StateAborted:
		m.mode = modeList
		m.status = theme.Warning.Render("cancelled")
		return m, nil
	}
	return m, cmd
}

func (m Model) applyIdentity(idx int, name, email string) tea.Cmd {
	env := m.env
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, kv := range [][2]string{{"user.name", name}, {"user.email", email}} {
			if r := env.Runner.Run(ctx, "git", "config", "--global", kv[0], kv[1]); !r.OK() {
				err := r.Err
				if err == nil {
					err = fmt.Errorf("git config %s: %s", kv[0], r.Combined())
				}
				return fixDoneMsg{idx: idx, err: err}
			}
		}
		return fixDoneMsg{idx: idx}
	}
}

// --- view ----------------------------------------------------------------------------

// View implements tea.Model.
func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	width := m.width
	if width <= 0 {
		width = 80
	}
	inner := clamp(width-2, 40, 110)

	var sections []string
	sections = append(sections, theme.Header("bootstrap", "Finish wiring this devcontainer: identity, logins and tooling."))
	sections = append(sections, m.viewProgress())

	if m.mode == modeForm {
		sections = append(sections, theme.Panel.Width(inner).Render(
			theme.Emph.Render(theme.GlyphBullet+" Git identity")+"\n\n"+m.form.View()))
		sections = append(sections, theme.Footer(theme.KeyHint("tab", "next field"), theme.KeyHint("enter", "confirm"), theme.KeyHint("esc", "cancel")))
	} else {
		sections = append(sections, m.viewSteps(inner))
		sections = append(sections, m.viewDetails(inner))
		sections = append(sections, m.viewFooter())
	}
	if m.status != "" {
		sections = append(sections, m.status)
	}

	content := lipgloss.NewStyle().Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, interleave(sections, "")...))
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "bootstrap"
	return v
}

func (m Model) viewProgress() string {
	done, total := m.requiredProgress()
	label := theme.Dim.Render(fmt.Sprintf("%d/%d required ready", done, total))
	if m.allRequiredReady() && m.noneChecking() {
		label = theme.Gradient("ALL SYSTEMS ONLINE", theme.Acid, theme.Cyan) + theme.Dim.Render("  optional steps below if you want them")
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.progress.ViewAs(m.percent()), label)
}

func (m Model) noneChecking() bool {
	for _, r := range m.results {
		if r.Status == StatusChecking {
			return false
		}
	}
	return true
}

func (m Model) viewSteps(width int) string {
	rows := make([]string, 0, len(m.steps)+1)
	for i, step := range m.steps {
		res := m.results[i]
		cursor := "  "
		title := theme.Text.Render(step.Title)
		if i == m.cursor {
			cursor = theme.Emph.Render(theme.GlyphCursor) + " "
			title = theme.Selected.Render(step.Title)
		}
		tag := ""
		if step.Optional {
			tag = " " + theme.Muted.Render("optional")
		}
		summary := ""
		if res.Summary != "" {
			summary = "  " + theme.Dim.Render(res.Summary)
		}
		line := cursor + m.glyph(i) + " " + title + tag + summary
		rows = append(rows, lipgloss.NewStyle().MaxWidth(width).Render(line))
	}
	return strings.Join(rows, "\n")
}

func (m Model) glyph(i int) string {
	switch m.results[i].Status {
	case StatusChecking:
		return m.spinner.View()
	case StatusReady:
		return theme.StatusGlyph("ready")
	case StatusNeeded:
		if m.steps[i].Optional {
			return theme.StatusGlyph("optional")
		}
		return theme.StatusGlyph("needed")
	case StatusSkipped:
		return theme.StatusGlyph("skipped")
	case StatusError:
		return theme.StatusGlyph("error")
	default:
		return theme.StatusGlyph("unknown")
	}
}

func (m Model) viewDetails(width int) string {
	step, res := m.steps[m.cursor], m.results[m.cursor]
	var lines []string
	lines = append(lines, theme.Emph.Render(theme.GlyphBullet+" "+step.Title))
	lines = append(lines, theme.Dim.Render(step.Description))
	if res.Detail != "" {
		lines = append(lines, "", theme.Text.Render(res.Detail))
	}
	switch {
	case res.Status == StatusReady:
		lines = append(lines, "", theme.Success.Render(theme.GlyphReady+" nothing to do"))
	case res.Status == StatusChecking:
		lines = append(lines, "", theme.Dim.Render("checking…"))
	case step.Fix.Kind == FixCommand:
		lines = append(lines, "", theme.Accent.Render(theme.GlyphArrow+" enter runs ")+theme.Code.Render(step.Fix.Label))
		if step.Fix.Note != "" {
			lines = append(lines, theme.Warning.Render(step.Fix.Note))
		}
	case step.Fix.Kind == FixGitIdentity:
		lines = append(lines, "", theme.Accent.Render(theme.GlyphArrow+" enter opens a short form"))
	}
	return theme.Panel.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) viewFooter() string {
	hints := []string{
		theme.KeyHint("↑↓", "move"),
		theme.KeyHint("enter", "fix"),
		theme.KeyHint("r", "re-check"),
	}
	if m.steps[m.cursor].Optional {
		hints = append(hints, theme.KeyHint("s", "skip"))
	}
	hints = append(hints, theme.KeyHint("q", "quit"))
	return theme.Footer(hints...)
}

// --- helpers ------------------------------------------------------------------------

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// interleave puts sep between every element of parts.
func interleave(parts []string, sep string) []string {
	out := make([]string, 0, len(parts)*2)
	for i, p := range parts {
		if i > 0 {
			out = append(out, sep)
		}
		out = append(out, p)
	}
	return out
}
