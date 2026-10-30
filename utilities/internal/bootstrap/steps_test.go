package bootstrap

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/bmorton/devcontainer/utilities/internal/sh"
)

func testEnv(results map[string]sh.Result, env map[string]string, files map[string]string) Env {
	return Env{
		Runner: &sh.FakeRunner{Results: results},
		Getenv: func(k string) string { return env[k] },
		Home:   "/home/test",
		ReadFile: func(p string) ([]byte, error) {
			if c, ok := files[p]; ok {
				return []byte(c), nil
			}
			return nil, os.ErrNotExist
		},
		FileOK:    func(p string) bool { _, ok := files[p]; return ok },
		Installed: func(string) bool { return true },
	}
}

func find(t *testing.T, id string) *Step {
	t.Helper()
	for _, s := range Steps() {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("no step %q", id)
	return nil
}

func TestGitIdentity(t *testing.T) {
	ctx := context.Background()
	step := find(t, "git-identity")

	t.Run("from env", func(t *testing.T) {
		env := testEnv(nil, map[string]string{"GIT_AUTHOR_NAME": "Ada", "GIT_AUTHOR_EMAIL": "ada@example.com"}, nil)
		r := step.Check(ctx, env)
		if r.Status != StatusReady || !strings.Contains(r.Summary, "Ada <ada@example.com>") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("from git config", func(t *testing.T) {
		env := testEnv(map[string]sh.Result{
			"git config --global --get user.name":  {Stdout: "Ada\n"},
			"git config --global --get user.email": {Stdout: "ada@example.com\n"},
		}, nil, nil)
		if r := step.Check(ctx, env); r.Status != StatusReady {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("missing email", func(t *testing.T) {
		env := testEnv(map[string]sh.Result{
			"git config --global --get user.name":  {Stdout: "Ada\n"},
			"git config --global --get user.email": {ExitCode: 1},
		}, nil, nil)
		r := step.Check(ctx, env)
		if r.Status != StatusNeeded || !strings.Contains(r.Summary, "user.email") || strings.Contains(r.Summary, "user.name") {
			t.Fatalf("got %+v", r)
		}
	})
}

func TestGitHubCLI(t *testing.T) {
	ctx := context.Background()
	step := find(t, "gh")
	out := "github.com\n  ✓ Logged in to github.com account bmorton (/home/x/.config/gh/hosts.yml)\n"
	env := testEnv(map[string]sh.Result{"gh auth status": {Stdout: out}}, nil, nil)
	r := step.Check(ctx, env)
	if r.Status != StatusReady || r.Summary != "logged in to github.com as bmorton" {
		t.Fatalf("got %+v", r)
	}
	env = testEnv(map[string]sh.Result{"gh auth status": {ExitCode: 1, Stderr: "You are not logged into any GitHub hosts."}}, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusNeeded {
		t.Fatalf("got %+v", r)
	}
}

func TestCredentialHelper(t *testing.T) {
	ctx := context.Background()
	step := find(t, "gh-git-credential")
	key := "git config --global --get-all credential.https://github.com.helper"

	env := testEnv(map[string]sh.Result{key: {Stdout: "\n!/usr/bin/gh auth git-credential\n"}}, nil, map[string]string{"/usr/bin/gh": ""})
	if r := step.Check(ctx, env); r.Status != StatusReady {
		t.Fatalf("existing helper: got %+v", r)
	}
	env = testEnv(map[string]sh.Result{key: {Stdout: "!/usr/bin/gh auth git-credential\n"}}, nil, nil)
	r := step.Check(ctx, env)
	if r.Status != StatusNeeded || !strings.Contains(r.Summary, "stale") {
		t.Fatalf("stale helper: got %+v", r)
	}
	env = testEnv(map[string]sh.Result{key: {ExitCode: 1}}, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusNeeded {
		t.Fatalf("no helper: got %+v", r)
	}
}

func TestClaude(t *testing.T) {
	ctx := context.Background()
	step := find(t, "claude")
	in := `{"loggedIn": true, "authMethod": "claude.ai", "apiProvider": "firstParty"}`
	env := testEnv(map[string]sh.Result{"claude auth status --json": {Stdout: in}}, nil, nil)
	r := step.Check(ctx, env)
	if r.Status != StatusReady || !strings.Contains(r.Summary, "claude.ai") {
		t.Fatalf("got %+v", r)
	}
	env = testEnv(map[string]sh.Result{"claude auth status --json": {Stdout: `{"loggedIn": false}`}}, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusNeeded {
		t.Fatalf("got %+v", r)
	}
}

func TestOpenCode(t *testing.T) {
	ctx := context.Background()
	step := find(t, "opencode")
	env := testEnv(map[string]sh.Result{"opencode auth list": {Stdout: "No authenticated integrations\n"}}, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusNeeded {
		t.Fatalf("got %+v", r)
	}
	env = testEnv(map[string]sh.Result{"opencode auth list": {Stdout: "anthropic oauth\nopenai api\n"}}, nil, nil)
	r := step.Check(ctx, env)
	if r.Status != StatusReady || r.Summary != "2 integration(s)" {
		t.Fatalf("got %+v", r)
	}
}

func TestPi(t *testing.T) {
	ctx := context.Background()
	step := find(t, "pi")
	auth := "/home/test/.pi/agent/auth.json"

	env := testEnv(nil, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusNeeded {
		t.Fatalf("no auth file: got %+v", r)
	}
	env = testEnv(map[string]sh.Result{
		"pi auth check --provider anthropic --json": {Stdout: `{"status":"not_ready","provider":"anthropic","reason":"credentials_not_configured"}`, ExitCode: 1},
		"pi auth check --provider synthetic --json": {Stdout: `{"status":"ready","provider":"synthetic"}`},
	}, nil, map[string]string{auth: `{"anthropic":{},"synthetic":{"type":"api_key"}}`})
	r := step.Check(ctx, env)
	if r.Status != StatusReady || r.Summary != "providers ready: synthetic" {
		t.Fatalf("got %+v", r)
	}
}

func TestMise(t *testing.T) {
	ctx := context.Background()
	step := find(t, "mise")
	env := testEnv(map[string]sh.Result{"mise ls --missing": {Stdout: ""}}, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusReady {
		t.Fatalf("got %+v", r)
	}
	env = testEnv(map[string]sh.Result{"mise ls --missing": {Stdout: "go  1.27.3  (missing)\n"}}, nil, nil)
	r := step.Check(ctx, env)
	if r.Status != StatusNeeded || !strings.Contains(r.Detail, "go") {
		t.Fatalf("got %+v", r)
	}
}

func TestRunnerErrorIsFailure(t *testing.T) {
	ctx := context.Background()
	step := find(t, "nsc")
	env := testEnv(map[string]sh.Result{"nsc auth check-login": {ExitCode: -1, Err: errors.New("boom")}}, nil, nil)
	if r := step.Check(ctx, env); r.Status != StatusError {
		t.Fatalf("got %+v", r)
	}
}

func TestStepsOrderRequiredFirst(t *testing.T) {
	seenOptional := false
	for _, s := range Steps() {
		if s.Optional {
			seenOptional = true
		} else if seenOptional {
			t.Fatalf("required step %q listed after an optional one", s.ID)
		}
	}
}
