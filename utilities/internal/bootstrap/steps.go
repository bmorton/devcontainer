package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Steps returns the ordered setup checklist. Required steps come first.
func Steps() []*Step {
	return []*Step{
		stepMise(),
		stepGitIdentity(),
		stepGitHubCLI(),
		stepGitCredentialHelper(),
		stepClaude(),
		stepCopilot(),
		stepOpenCode(),
		stepPi(),
		stepNamespace(),
		stepAzure(),
	}
}

func notInstalled(name string) Result {
	return Result{Status: StatusError, Summary: name + " is not installed", Detail: "Expected it on PATH via the mise shims. Run `mise install` and retry."}
}

// --- tooling ----------------------------------------------------------------

func stepMise() *Step {
	return &Step{
		ID:          "mise",
		Title:       "Toolchain",
		Description: "Every tool pinned in /etc/mise/config.toml is installed.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("mise") {
				return notInstalled("mise")
			}
			r := env.Runner.Run(ctx, "mise", "ls", "--missing")
			if !r.OK() {
				return Failed(fmt.Errorf("mise ls --missing: %s", r.Combined()))
			}
			if strings.TrimSpace(r.Stdout) == "" {
				return Ready("all pinned tools installed")
			}
			return Needed("some pinned tools are missing", strings.TrimSpace(r.Stdout))
		},
		Fix: Fix{Kind: FixCommand, Label: "mise install", Command: []string{"mise", "install"}},
	}
}

// --- git ---------------------------------------------------------------------

func gitConfig(ctx context.Context, env Env, key string) string {
	r := env.Runner.Run(ctx, "git", "config", "--global", "--get", key)
	if !r.OK() {
		return ""
	}
	return strings.TrimSpace(r.Stdout)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// identity returns the effective git author name and email.
func identity(ctx context.Context, env Env) (name, email string) {
	name = firstNonEmpty(env.Getenv("GIT_AUTHOR_NAME"), gitConfig(ctx, env, "user.name"))
	email = firstNonEmpty(env.Getenv("GIT_AUTHOR_EMAIL"), gitConfig(ctx, env, "user.email"))
	return name, email
}

func stepGitIdentity() *Step {
	return &Step{
		ID:          "git-identity",
		Title:       "Git identity",
		Description: "user.name and user.email so commits are attributed to you (Coder may inject these via GIT_AUTHOR_* variables).",
		Check: func(ctx context.Context, env Env) Result {
			name, email := identity(ctx, env)
			if name != "" && email != "" {
				return Ready(fmt.Sprintf("%s <%s>", name, email))
			}
			missing := []string{}
			if name == "" {
				missing = append(missing, "user.name")
			}
			if email == "" {
				missing = append(missing, "user.email")
			}
			return Needed("missing "+strings.Join(missing, " and "),
				"Press enter to fill in a short form, or run:\n  git config --global user.name \"Your Name\"\n  git config --global user.email you@example.com")
		},
		Fix: Fix{Kind: FixGitIdentity, Label: "set git identity"},
	}
}

var ghAccountRe = regexp.MustCompile(`Logged in to (\S+) account (\S+)`)

func stepGitHubCLI() *Step {
	return &Step{
		ID:          "gh",
		Title:       "GitHub CLI",
		Description: "gh is authenticated; it also feeds credentials to git and the Copilot CLI.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("gh") {
				return notInstalled("gh")
			}
			r := env.Runner.Run(ctx, "gh", "auth", "status")
			out := r.Combined()
			if r.Err != nil {
				return Failed(r.Err)
			}
			if r.ExitCode != 0 {
				return Needed("not logged in", out)
			}
			if m := ghAccountRe.FindStringSubmatch(out); m != nil {
				return Ready(fmt.Sprintf("logged in to %s as %s", m[1], m[2]))
			}
			return Ready("logged in")
		},
		Fix: Fix{Kind: FixCommand, Label: "gh auth login", Command: []string{"gh", "auth", "login"}},
	}
}

func stepGitCredentialHelper() *Step {
	return &Step{
		ID:          "gh-git-credential",
		Title:       "Git credential helper",
		Description: "git uses gh for https://github.com so clone/push work without prompts. A helper pointing at a gh binary that no longer exists (e.g. /usr/bin/gh from the old image) is flagged as stale.",
		Check: func(ctx context.Context, env Env) Result {
			r := env.Runner.Run(ctx, "git", "config", "--global", "--get-all", "credential.https://github.com.helper")
			if r.Err != nil {
				return Failed(r.Err)
			}
			var helper string
			for _, line := range strings.Split(r.Stdout, "\n") {
				if strings.Contains(line, "gh auth git-credential") {
					helper = strings.TrimSpace(line)
				}
			}
			if helper == "" {
				return Needed("gh is not configured as git's credential helper", "Run gh auth setup-git (after gh auth login).")
			}
			// "!/usr/bin/gh auth git-credential" -> /usr/bin/gh
			if strings.HasPrefix(helper, "!") {
				fields := strings.Fields(helper[1:])
				if len(fields) > 0 && filepath.IsAbs(fields[0]) && !env.FileOK(fields[0]) {
					return Needed("stale helper path "+fields[0], "The configured gh binary no longer exists. Re-run gh auth setup-git to point at the current gh.")
				}
			}
			return Ready("gh handles github.com credentials")
		},
		Fix: Fix{Kind: FixCommand, Label: "gh auth setup-git", Command: []string{"gh", "auth", "setup-git"}},
	}
}

// --- coding agents ------------------------------------------------------------

type claudeAuthStatus struct {
	LoggedIn   bool   `json:"loggedIn"`
	AuthMethod string `json:"authMethod"`
	Email      string `json:"email"`
}

func stepClaude() *Step {
	return &Step{
		ID:          "claude",
		Title:       "Claude Code",
		Description: "Signed in to your Anthropic account.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("claude") {
				return notInstalled("claude")
			}
			r := env.Runner.Run(ctx, "claude", "auth", "status", "--json")
			if r.Err != nil {
				return Failed(r.Err)
			}
			var st claudeAuthStatus
			if err := json.Unmarshal([]byte(strings.TrimSpace(r.Stdout)), &st); err != nil {
				if r.ExitCode != 0 {
					return Needed("not logged in", r.Combined())
				}
				return Failed(fmt.Errorf("parse claude auth status: %w", err))
			}
			if !st.LoggedIn {
				return Needed("not logged in", "")
			}
			via := st.AuthMethod
			if st.Email != "" {
				via = st.Email + " via " + via
			}
			return Ready("logged in (" + via + ")")
		},
		Fix: Fix{Kind: FixCommand, Label: "claude auth login", Command: []string{"claude", "auth", "login"}},
	}
}

var copilotTokenVars = []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"}

func stepCopilot() *Step {
	return &Step{
		ID:          "copilot",
		Title:       "GitHub Copilot CLI",
		Optional:    true,
		Description: "copilot authenticates from COPILOT_GITHUB_TOKEN / GH_TOKEN / GITHUB_TOKEN or a stored OAuth login. A stored login cannot be verified without launching copilot, so this step only proves the environment-variable path.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("copilot") {
				return notInstalled("copilot")
			}
			for _, v := range copilotTokenVars {
				if env.Getenv(v) != "" {
					return Ready("token from $" + v)
				}
			}
			return Needed("no token in environment; stored login unverified",
				"Press enter to run copilot login (device code), or export GH_TOKEN=\"$(gh auth token)\".")
		},
		Fix: Fix{Kind: FixCommand, Label: "copilot login --device-code", Command: []string{"copilot", "login", "--device-code"}, Unverifiable: true},
	}
}

func stepOpenCode() *Step {
	return &Step{
		ID:          "opencode",
		Title:       "opencode",
		Optional:    true,
		Description: "At least one model provider is authenticated.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("opencode") {
				return notInstalled("opencode")
			}
			r := env.Runner.Run(ctx, "opencode", "auth", "list")
			if r.Err != nil {
				return Failed(r.Err)
			}
			out := strings.TrimSpace(r.Combined())
			if r.ExitCode != 0 {
				return Failed(fmt.Errorf("opencode auth list: %s", out))
			}
			if out == "" || strings.Contains(out, "No authenticated integrations") {
				return Needed("no providers authenticated", "")
			}
			n := 0
			for _, l := range strings.Split(out, "\n") {
				if strings.TrimSpace(l) != "" {
					n++
				}
			}
			return Ready(fmt.Sprintf("%d integration(s)", n))
		},
		Fix: Fix{Kind: FixCommand, Label: "opencode auth login", Command: []string{"opencode", "auth", "login"}},
	}
}

type piAuthCheck struct {
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Reason   string `json:"reason"`
}

func stepPi() *Step {
	return &Step{
		ID:          "pi",
		Title:       "pi",
		Optional:    true,
		Description: "pi has credentials for at least one model provider (~/.pi/agent/auth.json). Login happens inside pi with /login.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("pi") {
				return notInstalled("pi")
			}
			raw, err := env.ReadFile(filepath.Join(env.Home, ".pi", "agent", "auth.json"))
			if err != nil {
				return Needed("no provider credentials yet", "")
			}
			var providers map[string]json.RawMessage
			if err := json.Unmarshal(raw, &providers); err != nil || len(providers) == 0 {
				return Needed("no provider credentials yet", "")
			}
			names := make([]string, 0, len(providers))
			for p := range providers {
				names = append(names, p)
			}
			sort.Strings(names)
			var ready, notReady []string
			for _, p := range names {
				r := env.Runner.Run(ctx, "pi", "auth", "check", "--provider", p, "--json")
				var st piAuthCheck
				if json.Unmarshal([]byte(strings.TrimSpace(r.Stdout)), &st) == nil && st.Status == "ready" {
					ready = append(ready, p)
				} else {
					notReady = append(notReady, p)
				}
			}
			if len(ready) > 0 {
				return Ready("providers ready: " + strings.Join(ready, ", "))
			}
			return Needed("configured providers are not ready: "+strings.Join(notReady, ", "), "")
		},
		Fix: Fix{Kind: FixCommand, Label: "pi", Command: []string{"pi"},
			Note: "pi will open. Run /login to add a provider, then quit pi to return here."},
	}
}

// --- platform CLIs --------------------------------------------------------------

func stepNamespace() *Step {
	return &Step{
		ID:          "nsc",
		Title:       "Namespace (nsc)",
		Optional:    true,
		Description: "Logged in to Namespace for remote builders and instances.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("nsc") {
				return notInstalled("nsc")
			}
			r := env.Runner.Run(ctx, "nsc", "auth", "check-login")
			if r.Err != nil {
				return Failed(r.Err)
			}
			if r.ExitCode != 0 {
				return Needed("not logged in", "")
			}
			return Ready("authenticated")
		},
		Fix: Fix{Kind: FixCommand, Label: "nsc login", Command: []string{"nsc", "login"}},
	}
}

type azAccount struct {
	Name string `json:"name"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
}

func stepAzure() *Step {
	return &Step{
		ID:          "az",
		Title:       "Azure CLI",
		Optional:    true,
		Description: "az has an active subscription.",
		Check: func(ctx context.Context, env Env) Result {
			if !env.Installed("az") {
				return notInstalled("az")
			}
			r := env.Runner.Run(ctx, "az", "account", "show", "-o", "json")
			if r.Err != nil {
				return Failed(r.Err)
			}
			if r.ExitCode != 0 {
				return Needed("not logged in", "")
			}
			var acct azAccount
			if json.Unmarshal([]byte(r.Stdout), &acct) == nil && acct.User.Name != "" {
				return Ready(fmt.Sprintf("%s (%s)", acct.User.Name, acct.Name))
			}
			return Ready("logged in")
		},
		Fix: Fix{Kind: FixCommand, Label: "az login --use-device-code", Command: []string{"az", "login", "--use-device-code"}},
	}
}
