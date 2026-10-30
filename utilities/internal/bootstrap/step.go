// Package bootstrap checks and completes the per-container setup that an image
// cannot bake in: logins, identity and tool health. It exposes a Bubble Tea
// TUI (Run) and a non-interactive report (RunCheck).
package bootstrap

import (
	"context"
	"os"

	"github.com/bmorton/devcontainer/utilities/internal/sh"
)

// Status is the outcome of a Step's check.
type Status int

const (
	StatusUnknown  Status = iota // not checked yet
	StatusChecking               // check in flight
	StatusReady                  // nothing to do
	StatusNeeded                 // action required
	StatusSkipped                // optional step the user chose to skip
	StatusError                  // the check itself failed (tool missing, timeout)
)

// Result is what a check reports back.
type Result struct {
	Status  Status
	Summary string // one line, shown next to the step title
	Detail  string // optional multi-line detail shown in the panel
}

// Ready builds a StatusReady result.
func Ready(summary string) Result { return Result{Status: StatusReady, Summary: summary} }

// Needed builds a StatusNeeded result.
func Needed(summary, detail string) Result {
	return Result{Status: StatusNeeded, Summary: summary, Detail: detail}
}

// Failed builds a StatusError result from an error.
func Failed(err error) Result {
	return Result{Status: StatusError, Summary: "check failed", Detail: err.Error()}
}

// FixKind says how a step is completed.
type FixKind int

const (
	FixNone        FixKind = iota // manual; Detail explains what to do
	FixCommand                    // run Command interactively in the terminal
	FixGitIdentity                // in-TUI form that writes git config
)

// Fix describes the action offered for a StatusNeeded step.
type Fix struct {
	Kind    FixKind
	Label   string   // what we tell the user we're about to run
	Command []string // for FixCommand
	Note    string   // guidance shown before running (e.g. "use /login inside pi")
	// Unverifiable marks fixes whose success cannot be re-checked afterwards;
	// a clean exit is recorded as ready instead of re-running the check.
	Unverifiable bool
}

// Env is everything a check may touch, injected so tests can fake it.
type Env struct {
	Runner   sh.Runner
	Getenv   func(string) string
	Home     string
	ReadFile func(string) ([]byte, error)
	FileOK   func(string) bool // exists and is readable
	// Installed reports whether a command resolves on PATH.
	Installed func(string) bool
}

// DefaultEnv wires Env to the real process environment.
func DefaultEnv() Env {
	home, _ := os.UserHomeDir()
	return Env{
		Runner:   sh.ExecRunner{},
		Getenv:   os.Getenv,
		Home:     home,
		ReadFile: os.ReadFile,
		FileOK: func(p string) bool {
			_, err := os.Stat(p)
			return err == nil
		},
		Installed: sh.Installed,
	}
}

// Step is one unit of setup.
type Step struct {
	ID          string
	Title       string
	Description string
	Optional    bool
	Check       func(ctx context.Context, env Env) Result
	Fix         Fix
}

// Required is the inverse of Optional, for readability at call sites.
func (s *Step) Required() bool { return !s.Optional }
