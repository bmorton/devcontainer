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
#   3. the lifecycle ran to completion (postCreateCommand sentinel exists).
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

echo "=== Verification complete ==="
if [ "$failures" -ne 0 ]; then
  echo "$failures check(s) failed." >&2
  exit 1
fi
echo "All checks passed."
