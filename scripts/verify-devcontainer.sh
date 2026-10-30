#!/usr/bin/env bash
#
# Verify that the built dev container started up correctly.
#
# This runs *inside* the container after the build and lifecycle commands have
# completed. It is used both as the envbuilder `ENVBUILDER_INIT_SCRIPT` (the
# source of truth for how Coder builds this workspace) and can be run by hand
# inside any container started from this dev container.
#
# It asserts the things that have previously broken on Coder/envbuilder:
#   1. the effective user is `node`, not root,
#   2. tmux is on PATH and can start a session that reads ~/.tmux.conf,
#   3. the lifecycle ran to completion (postCreateCommand sentinel exists),
#   4. /workspaces is writable by the user (git clone used to fail there),
#   5. every tool pinned in mise.toml is installed and resolvable via shims.
#
# Any failed check exits non-zero so CI fails loudly instead of silently
# dropping you into a broken shell.

set -euo pipefail

failures=0

fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

pass() {
  echo "PASS: $*"
}

echo "=== Verifying dev container startup ==="
echo "whoami: $(whoami)"
echo "id: $(id)"
echo "HOME: ${HOME:-<unset>}"

# 1. Effective user must be `node`, not root.
if [ "$(id -un)" = "node" ]; then
  pass "running as expected user 'node'"
else
  fail "expected to run as 'node' but running as '$(id -un)' (uid=$(id -u))"
fi

# 2. tmux must be on PATH and able to start a session that reads ~/.tmux.conf.
if command -v tmux >/dev/null 2>&1; then
  pass "tmux is on PATH ($(command -v tmux))"

  session="verify-$$"
  tmux_conf="${HOME}/.tmux.conf"
  tmux_err="$(mktemp)"

  if [ -f "$tmux_conf" ]; then
    pass "tmux config present at $tmux_conf"
  else
    fail "tmux config missing at $tmux_conf"
  fi

  # Start a detached session, explicitly loading the user's config so a broken
  # config (or a missing HOME) fails the check rather than silently starting a
  # default session.
  if tmux -f "$tmux_conf" new-session -d -s "$session" 'sleep 5' 2>"$tmux_err"; then
    if tmux has-session -t "$session" 2>/dev/null; then
      pass "tmux started a session using $tmux_conf"
    else
      fail "tmux session '$session' did not stay alive"
    fi
    tmux kill-session -t "$session" 2>/dev/null || true
  else
    fail "tmux could not start a session: $(cat "$tmux_err" 2>/dev/null)"
  fi
  rm -f "$tmux_err"
else
  fail "tmux is not on PATH"
fi

# 3. The lifecycle (postCreateCommand) must have completed.
sentinel="${HOME}/.devcontainer-postcreate-done"
if [ -f "$sentinel" ]; then
  pass "lifecycle sentinel present ($sentinel)"
else
  fail "lifecycle sentinel missing ($sentinel); postCreateCommand did not complete"
fi

# 4. /workspaces must be writable by the current user (envbuilder leaves the
#    base image's root-owned directory alone; the Dockerfile chowns it).
probe="/workspaces/.verify-write-$$"
if mkdir "$probe" 2>/dev/null; then
  rmdir "$probe"
  pass "/workspaces is writable by $(id -un)"
else
  fail "/workspaces is not writable by $(id -un) (owner: $(stat -c '%U:%G' /workspaces 2>/dev/null || echo unknown))"
fi

# 5. mise-managed tools: nothing pinned in /etc/mise/config.toml may be missing,
#    and a representative set of shims must execute from a plain PATH lookup
#    (no `mise activate`), which is how VS Code / Coder / CI shells see them.
if command -v mise >/dev/null 2>&1; then
  pass "mise is on PATH ($(mise --version 2>/dev/null | head -1))"
  missing="$(mise ls --missing 2>/dev/null || true)"
  if [ -z "$missing" ]; then
    pass "all mise tools from the system config are installed"
  else
    fail "mise reports missing tools:"$'\n'"$missing"
  fi
  for tool in go ruby cargo kubectl helm gh az claude opencode copilot pi witr tmux gopls golangci-lint; do
    if command -v "$tool" >/dev/null 2>&1 && "$tool" --version >/dev/null 2>&1; then
      pass "$tool runs via shim ($(command -v "$tool"))"
    else
      fail "$tool is missing or fails to run"
    fi
  done
  if [ -x /usr/local/bin/nsc ]; then
    pass "nsc installed"
  else
    fail "nsc missing from /usr/local/bin"
  fi
else
  fail "mise is not on PATH"
fi

echo "=== Verification complete ==="
if [ "$failures" -ne 0 ]; then
  echo "$failures check(s) failed." >&2
  exit 1
fi
echo "All checks passed."
