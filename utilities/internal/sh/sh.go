// Package sh runs external commands behind a small interface so callers can be
// unit-tested with a fake runner instead of a real shell.
package sh

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
)

// Result is the outcome of running a command to completion.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	// Err is set only when the command could not be started (not found,
	// permission denied, context cancelled). A non-zero exit is not an Err.
	Err error
}

// OK reports whether the command started and exited 0.
func (r Result) OK() bool { return r.Err == nil && r.ExitCode == 0 }

// Combined returns stdout followed by stderr, trimmed.
func (r Result) Combined() string {
	return strings.TrimSpace(r.Stdout + "\n" + r.Stderr)
}

// Runner executes a command non-interactively (no stdin) and captures output.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) Result
}

// ExecRunner is the real Runner backed by os/exec.
type ExecRunner struct{}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, name string, args ...string) Result {
	path, err := exec.LookPath(name)
	if err != nil {
		return Result{ExitCode: 127, Err: err}
	}
	cmd := exec.CommandContext(ctx, path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	res := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
			res.Err = err
		}
	}
	return res
}

// Installed reports whether name resolves on PATH.
func Installed(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// FakeRunner is a Runner for tests. Results are keyed by the full command
// line ("gh auth status"); unmatched commands return exit 127.
type FakeRunner struct {
	Results map[string]Result
	Calls   []string
}

// Run implements Runner.
func (f *FakeRunner) Run(_ context.Context, name string, args ...string) Result {
	key := strings.TrimSpace(name + " " + strings.Join(args, " "))
	f.Calls = append(f.Calls, key)
	if r, ok := f.Results[key]; ok {
		return r
	}
	return Result{ExitCode: 127, Err: errors.New("fake: no result for " + key)}
}
