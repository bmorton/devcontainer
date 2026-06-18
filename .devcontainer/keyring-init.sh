#!/usr/bin/env bash
# Bootstraps a per-container Linux Secret Service so VS Code's
# SecretStorage / keytar / libsecret consumers (incl. GitHub Copilot)
# have a local encrypted credential vault.
#
# Idempotent: safe to run on every container start and VS Code attach.
# Run as the remote user; do NOT run as root.

set -euo pipefail

log()  { printf '[keyring-init] %s\n' "$*"; }
warn() { printf '[keyring-init] WARNING: %s\n' "$*" >&2; }

UID_NUM="$(id -u)"
XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/tmp/runtime-${UID_NUM}}"
mkdir -p "$XDG_RUNTIME_DIR"
chmod 700 "$XDG_RUNTIME_DIR"
export XDG_RUNTIME_DIR

DBUS_SOCKET="${XDG_RUNTIME_DIR}/bus"
DBUS_PID_FILE="${XDG_RUNTIME_DIR}/dbus.pid"
ENV_FILE="${XDG_RUNTIME_DIR}/keyring.env"

# --- 1. DBus session bus on a stable socket path --------------------------
dbus_alive() {
    [[ -S "$DBUS_SOCKET" ]] || return 1
    [[ -f "$DBUS_PID_FILE" ]] || return 1
    local pid; pid="$(cat "$DBUS_PID_FILE" 2>/dev/null || true)"
    [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

if dbus_alive; then
    log "reusing existing dbus-daemon on $DBUS_SOCKET"
else
    rm -f "$DBUS_SOCKET" "$DBUS_PID_FILE"
    if ! dbus-daemon --session --address="unix:path=${DBUS_SOCKET}" \
            --nopidfile --fork --print-pid=3 3>"$DBUS_PID_FILE"; then
        warn "failed to start dbus-daemon"; exit 0
    fi
    for _ in 1 2 3 4 5 6 7 8 9 10; do
        [[ -S "$DBUS_SOCKET" ]] && break
        sleep 0.1
    done
    [[ -S "$DBUS_SOCKET" ]] || { warn "dbus socket never appeared"; exit 0; }
    log "started dbus-daemon (pid $(cat "$DBUS_PID_FILE")) on $DBUS_SOCKET"
fi
export DBUS_SESSION_BUS_ADDRESS="unix:path=${DBUS_SOCKET}"

# --- 2. gnome-keyring-daemon ----------------------------------------------
keyring_alive() {
    pgrep -u "$UID_NUM" -x gnome-keyring-d >/dev/null 2>&1
}

if keyring_alive; then
    log "reusing existing gnome-keyring-daemon"
else
    KEYRING_PASS="${KEYRING_PASSWORD-}"
    gk_env="$(printf '%s' "$KEYRING_PASS" \
        | gnome-keyring-daemon --daemonize --unlock --components=secrets 2>/dev/null)" || {
        warn "gnome-keyring-daemon failed to start"; exit 0
    }
    log "started gnome-keyring-daemon"
    printf '%s\n' "$gk_env" > "${XDG_RUNTIME_DIR}/keyring.daemon-env"
fi

# --- 3. Bootstrap the persistent default collection (first run only) ------
if command -v python3 >/dev/null 2>&1 \
   && python3 -c 'import jeepney' >/dev/null 2>&1; then
    KEYRING_PASSWORD="${KEYRING_PASSWORD-}" \
        python3 "$(dirname "$0")/keyring-bootstrap.py" \
        || warn "default-collection bootstrap failed; SecretStorage may be read-only"
else
    warn "python3-jeepney not available; cannot bootstrap default collection"
fi

# --- 4. Persist env for non-VS-Code shells --------------------------------
{
    echo "export XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR}"
    echo "export DBUS_SESSION_BUS_ADDRESS=${DBUS_SESSION_BUS_ADDRESS}"
    [[ -f "${XDG_RUNTIME_DIR}/keyring.daemon-env" ]] \
        && sed -n 's/^\([A-Z_]\+\)=\(.*\)$/export \1=\2/p' \
               "${XDG_RUNTIME_DIR}/keyring.daemon-env"
} > "$ENV_FILE"

# --- 5. Verify Secret Service round-trip ----------------------------------
if command -v secret-tool >/dev/null 2>&1; then
    probe_value="ok-$$"
    if printf '%s' "$probe_value" \
            | secret-tool store --label=devcontainer-keyring-probe \
                  app devcontainer-keyring-probe 2>/dev/null \
       && [[ "$(secret-tool lookup app devcontainer-keyring-probe 2>/dev/null)" == "$probe_value" ]]; then
        secret-tool clear app devcontainer-keyring-probe 2>/dev/null || true
        log "Secret Service verified on $DBUS_SESSION_BUS_ADDRESS"
    else
        warn "Secret Service probe failed; VS Code extensions may fall back to the plaintext basic store"
    fi
fi
