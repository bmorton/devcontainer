package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
)

// Options controls Run.
type Options struct {
	// Check forces the non-interactive report even on a TTY.
	Check bool
	// Out receives the report in check mode (defaults to stdout).
	Out io.Writer
}

// Run is the program entrypoint shared by the cmd package. It returns the
// process exit code.
func Run(opts Options) int {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	env := DefaultEnv()
	steps := Steps()

	if opts.Check || !isTerminal(os.Stdout) || !isTerminal(os.Stdin) {
		return RunCheck(context.Background(), env, steps, opts.Out)
	}

	p := tea.NewProgram(NewModel(env, steps))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap:", err)
		return 1
	}
	return 0
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
