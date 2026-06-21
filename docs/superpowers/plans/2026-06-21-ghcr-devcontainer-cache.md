# GHCR Devcontainer Cache Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and publish a fully cached devcontainer image to a public GHCR repo from GitHub Actions, so Coder's envbuilder boots workspaces from the cache instead of rebuilding from scratch.

**Architecture:** A new GitHub Actions workflow runs the *same* envbuilder Coder uses (`docker run` with the repo mounted), configured with `ENVBUILDER_CACHE_REPO` + `ENVBUILDER_PUSH_IMAGE=1` to push every layer and a complete prebuilt image to `ghcr.io/bmorton/devcontainer-cache`. A thin `scripts/build-cache.sh` wraps the `docker run` invocation (mirroring the existing `scripts/test-envbuilder.sh`). Coder sets the same `ENVBUILDER_CACHE_REPO` to consume the cache. The build is verified by reusing `scripts/verify-devcontainer.sh` as the envbuilder init script.

**Tech Stack:** Bash, GitHub Actions, envbuilder (`ghcr.io/coder/envbuilder`), Docker, GHCR.

## Global Constraints

- envbuilder image pinned to `ghcr.io/coder/envbuilder:1.3.0` (exact, verbatim) — must be identical in CI and on the Coder side for layer hashes to match.
- Cache repo: `ghcr.io/bmorton/devcontainer-cache` (public; lowercase owner/name).
- Runner: `namespace-profile-devcontainer` (same as existing jobs; has Docker).
- Architecture: amd64 only.
- Workspace folder inside envbuilder: `/workspaces/devcontainer` (matches `scripts/test-envbuilder.sh`).
- Never print registry credentials (`ENVBUILDER_DOCKER_CONFIG_BASE64`) to logs.
- `ENVBUILDER_CACHE_TTL_DAYS=30` default in the cache build (envbuilder default is 7; longer avoids expiry between weekly runs).

---

### Task 1: `scripts/build-cache.sh` publishing script

**Files:**
- Create: `scripts/build-cache.sh`
- Test: `scripts/build-cache_test.sh`

**Interfaces:**
- Consumes (env): `CACHE_REPO` (required), `ENVBUILDER_DOCKER_CONFIG_BASE64` (required), `ENVBUILDER_IMAGE` (optional override, default `ghcr.io/coder/envbuilder:1.3.0`), `ENVBUILDER_CACHE_TTL_DAYS` (optional, default `30`), `DRY_RUN` (optional; when set, prints a redacted config summary and exits 0 without requiring Docker).
- Produces: an executable script the workflow calls as `scripts/build-cache.sh`. On a real run it `exec`s `docker run` with envbuilder configured to build + push the cache. DRY_RUN output lines (exact prefixes other code/tests rely on): `DRY_RUN: envbuilder_image=...`, `DRY_RUN: cache_repo=...`, `DRY_RUN: push_image=1 ...`, `DRY_RUN: init_script=...`, `DRY_RUN: docker_config_base64=<redacted ...>`.

- [ ] **Step 1: Write the failing test**

Create `scripts/build-cache_test.sh`:

```bash
#!/usr/bin/env bash
# Tests for scripts/build-cache.sh: env validation + DRY_RUN behavior.
# Runs without Docker (DRY_RUN path). Usage: bash scripts/build-cache_test.sh
set -uo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
target="${script_dir}/build-cache.sh"
failures=0

check() { # check "desc" cmd...
  local desc="$1"; shift
  if "$@"; then echo "PASS: ${desc}"; else echo "FAIL: ${desc}" >&2; failures=$((failures + 1)); fi
}

# 1. Fails when CACHE_REPO is unset/empty.
out=$(CACHE_REPO="" ENVBUILDER_DOCKER_CONFIG_BASE64="x" bash "$target" 2>&1); rc=$?
check "exits non-zero without CACHE_REPO" test "$rc" -ne 0
check "error mentions CACHE_REPO" grep -q "CACHE_REPO" <<<"$out"

# 2. Fails when registry credentials are unset/empty.
out=$(CACHE_REPO="ghcr.io/x/y" ENVBUILDER_DOCKER_CONFIG_BASE64="" bash "$target" 2>&1); rc=$?
check "exits non-zero without creds" test "$rc" -ne 0
check "error mentions ENVBUILDER_DOCKER_CONFIG_BASE64" grep -q "ENVBUILDER_DOCKER_CONFIG_BASE64" <<<"$out"

# 3. DRY_RUN prints redacted config and exits 0 without Docker.
out=$(CACHE_REPO="ghcr.io/bmorton/devcontainer-cache" \
      ENVBUILDER_DOCKER_CONFIG_BASE64="ZHVtbXk=" DRY_RUN=1 bash "$target" 2>&1); rc=$?
check "DRY_RUN exits 0" test "$rc" -eq 0
check "DRY_RUN shows cache_repo" grep -q "cache_repo=ghcr.io/bmorton/devcontainer-cache" <<<"$out"
check "DRY_RUN shows push_image=1" grep -q "push_image=1" <<<"$out"
check "DRY_RUN pins envbuilder 1.3.0" grep -q "envbuilder:1.3.0" <<<"$out"
check "DRY_RUN references verify script" grep -q "verify-devcontainer.sh" <<<"$out"
check "DRY_RUN redacts creds label" grep -q "redacted" <<<"$out"
if grep -q "ZHVtbXk=" <<<"$out"; then
  echo "FAIL: DRY_RUN leaked credentials" >&2; failures=$((failures + 1))
else
  echo "PASS: DRY_RUN did not leak credentials"
fi

if [ "$failures" -ne 0 ]; then echo "${failures} check(s) failed." >&2; exit 1; fi
echo "All build-cache tests passed."
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `bash scripts/build-cache_test.sh`
Expected: FAIL — `scripts/build-cache.sh` does not exist yet, so `bash "$target"` errors and assertions fail (non-zero exit).

- [ ] **Step 3: Write the script**

Create `scripts/build-cache.sh`:

```bash
#!/usr/bin/env bash
#
# Build this dev container with envbuilder (the builder Coder uses) and push a
# fully cached image to a container registry, so Coder workspaces start fast by
# pulling cached layers / the prebuilt image instead of rebuilding from scratch.
#
# This mirrors scripts/test-envbuilder.sh (repo mounted into envbuilder,
# .devcontainer/devcontainer.json auto-discovered) but is configured to PUBLISH
# the cache. It reuses scripts/verify-devcontainer.sh as the init script so a
# broken build (e.g. a base-image regression on the scheduled run) fails loudly.
#
# Usage (normally invoked by .github/workflows/cache-devcontainer.yml):
#   CACHE_REPO=ghcr.io/bmorton/devcontainer-cache \
#   ENVBUILDER_DOCKER_CONFIG_BASE64=<base64 docker config.json> \
#   scripts/build-cache.sh
#
# Required environment variables:
#   CACHE_REPO                      registry repo to push the cache to
#   ENVBUILDER_DOCKER_CONFIG_BASE64 base64-encoded docker config.json with creds
#
# Optional environment variables:
#   ENVBUILDER_IMAGE          envbuilder image (default: ghcr.io/coder/envbuilder:1.3.0)
#   ENVBUILDER_CACHE_TTL_DAYS cache layer TTL in days (default: 30)
#   DRY_RUN                   if set, print a redacted config summary and exit 0
#                             without requiring Docker (used by tests)

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
envbuilder_image="${ENVBUILDER_IMAGE:-ghcr.io/coder/envbuilder:1.3.0}"
cache_repo="${CACHE_REPO:-}"
cache_ttl_days="${ENVBUILDER_CACHE_TTL_DAYS:-30}"
workspace_folder="/workspaces/devcontainer"
init_script="bash ${workspace_folder}/scripts/verify-devcontainer.sh"

if [ -z "$cache_repo" ]; then
  echo "error: CACHE_REPO is required (e.g. ghcr.io/bmorton/devcontainer-cache)" >&2
  exit 1
fi

if [ -z "${ENVBUILDER_DOCKER_CONFIG_BASE64:-}" ]; then
  echo "error: ENVBUILDER_DOCKER_CONFIG_BASE64 is required (base64 docker config.json)" >&2
  exit 1
fi

docker_args=(
  run --rm
  -e "ENVBUILDER_WORKSPACE_FOLDER=${workspace_folder}"
  -e "ENVBUILDER_INIT_SCRIPT=${init_script}"
  -e "ENVBUILDER_CACHE_REPO=${cache_repo}"
  -e "ENVBUILDER_PUSH_IMAGE=1"
  -e "ENVBUILDER_EXIT_ON_PUSH_FAILURE=1"
  -e "ENVBUILDER_CACHE_TTL_DAYS=${cache_ttl_days}"
  -e "ENVBUILDER_DOCKER_CONFIG_BASE64=${ENVBUILDER_DOCKER_CONFIG_BASE64}"
  -v "${repo_root}:${workspace_folder}"
  "${envbuilder_image}"
)

echo "=== Building + pushing devcontainer cache ==="
echo "DRY_RUN: envbuilder_image=${envbuilder_image}"
echo "DRY_RUN: cache_repo=${cache_repo}"
echo "DRY_RUN: push_image=1 exit_on_push_failure=1 cache_ttl_days=${cache_ttl_days}"
echo "DRY_RUN: init_script=${init_script}"
echo "DRY_RUN: docker_config_base64=<redacted, ${#ENVBUILDER_DOCKER_CONFIG_BASE64} bytes>"

if [ -n "${DRY_RUN:-}" ]; then
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker is required but not found on PATH" >&2
  exit 1
fi

exec docker "${docker_args[@]}"
```

Then make both files executable:

```bash
chmod +x scripts/build-cache.sh scripts/build-cache_test.sh
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `bash scripts/build-cache_test.sh`
Expected: PASS — final line `All build-cache tests passed.` and exit 0.

- [ ] **Step 5: Lint the script (if shellcheck available)**

Run: `command -v shellcheck >/dev/null && shellcheck scripts/build-cache.sh scripts/build-cache_test.sh || echo "shellcheck not installed; skipping"`
Expected: no warnings, or the "skipping" message. Fix any reported issues.

- [ ] **Step 6: Commit**

```bash
git add scripts/build-cache.sh scripts/build-cache_test.sh
git commit -m "Add build-cache.sh to publish envbuilder cache to GHCR

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

### Task 2: `cache-devcontainer.yml` workflow

**Files:**
- Create: `.github/workflows/cache-devcontainer.yml`

**Interfaces:**
- Consumes: `scripts/build-cache.sh` (Task 1) and its env contract (`CACHE_REPO`, `ENVBUILDER_DOCKER_CONFIG_BASE64`, `ENVBUILDER_IMAGE`).
- Produces: a workflow that, on `push` to `main` / weekly `schedule` / `workflow_dispatch`, builds and pushes the cache to `ghcr.io/bmorton/devcontainer-cache`.

- [ ] **Step 1: Write the workflow file**

Create `.github/workflows/cache-devcontainer.yml`:

```yaml
name: Cache Devcontainer

# Build the dev container with envbuilder (the builder Coder uses) and push a
# fully cached image to GHCR, so Coder workspaces start fast by reusing the cache
# instead of rebuilding from scratch. See scripts/build-cache.sh.
on:
  push:
    branches: [ main ]
  schedule:
    # Weekly (Mondays 06:00 UTC) to pick up base-image / feature security updates.
    - cron: '0 6 * * 1'
  workflow_dispatch:

permissions:
  contents: read
  packages: write

jobs:
  cache-devcontainer:
    name: Build and push devcontainer cache
    runs-on: namespace-profile-devcontainer
    timeout-minutes: 60

    env:
      # Pinned envbuilder version: MUST match the version Coder runs, or cached
      # layer hashes will not match and the cache will be ignored.
      ENVBUILDER_IMAGE: ghcr.io/coder/envbuilder:1.3.0
      CACHE_REPO: ghcr.io/bmorton/devcontainer-cache

    steps:
    - name: Checkout code
      uses: actions/checkout@v6

    - name: Build and push devcontainer cache
      env:
        GHCR_USER: ${{ github.actor }}
        GHCR_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: |
        # Build a base64-encoded docker config.json for GHCR and hand it to
        # envbuilder via ENVBUILDER_DOCKER_CONFIG_BASE64. Kept in-step (never
        # written to $GITHUB_ENV) so the credential is not persisted.
        auth=$(printf '%s:%s' "$GHCR_USER" "$GHCR_TOKEN" | base64 -w0)
        export ENVBUILDER_DOCKER_CONFIG_BASE64=$(printf '{"auths":{"ghcr.io":{"auth":"%s"}}}' "$auth" | base64 -w0)
        scripts/build-cache.sh
```

- [ ] **Step 2: Validate YAML parses**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/cache-devcontainer.yml')); print('yaml ok')"`
Expected: `yaml ok` (exit 0). Note: PyYAML parses the `on:` key as the boolean `True` — that is expected and fine; GitHub Actions reads it correctly.

- [ ] **Step 3: Validate structural requirements**

Run:
```bash
f=.github/workflows/cache-devcontainer.yml
for needle in "workflow_dispatch" "schedule" "cron: '0 6 * * 1'" "branches: [ main ]" \
  "packages: write" "namespace-profile-devcontainer" \
  "ENVBUILDER_IMAGE: ghcr.io/coder/envbuilder:1.3.0" \
  "ghcr.io/bmorton/devcontainer-cache" "scripts/build-cache.sh"; do
  grep -qF "$needle" "$f" && echo "ok: $needle" || { echo "MISSING: $needle" >&2; exit 1; }
done
```
Expected: an `ok:` line for each needle, exit 0.

- [ ] **Step 4: Lint with actionlint (if available)**

Run: `command -v actionlint >/dev/null && actionlint .github/workflows/cache-devcontainer.yml || echo "actionlint not installed; skipping"`
Expected: no errors, or the "skipping" message.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/cache-devcontainer.yml
git commit -m "Add workflow to build and push devcontainer cache to GHCR

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

### Task 3: README — Coder consumption + GHCR cache docs

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: the workflow (Task 2) and cache repo `ghcr.io/bmorton/devcontainer-cache`.
- Produces: documentation; no code consumers.

- [ ] **Step 1: Add the documentation section**

Append the following section to the end of `README.md`:

```markdown
## Faster Coder startup: prebuilt image cache on GHCR

Coder builds this dev container with [envbuilder](https://github.com/coder/envbuilder),
which otherwise rebuilds the whole image (Dockerfile **and** every feature —
Go, Ruby, Rust, kubectl/helm/minikube, Azure CLI, Playwright, …) on every
workspace start.

The [`Cache Devcontainer`](.github/workflows/cache-devcontainer.yml) workflow
runs the same envbuilder and pushes a fully cached image to
`ghcr.io/bmorton/devcontainer-cache` on every push to `main`, weekly, and on
manual dispatch. Coder then reuses the cache instead of rebuilding.

### One-time setup

After the first successful run, make the GHCR package **public** so Coder can
pull it without credentials: GitHub → Packages → `devcontainer-cache` → Package
settings → Change visibility → Public.

### Coder workspace configuration

Configure the workspace template's envbuilder with **the same pinned envbuilder
version** as CI (`ghcr.io/coder/envbuilder:1.3.0`) and:

| Environment variable | Value |
| --- | --- |
| `ENVBUILDER_CACHE_REPO` | `ghcr.io/bmorton/devcontainer-cache` |
| `ENVBUILDER_GET_CACHED_IMAGE` | `true` (optional) — boot directly from the prebuilt image, the fastest path |

The cache repo is public, so no pull credentials are needed.

### Why the version must match

Cache hits require identical build inputs and tooling between CI and Coder:
the **same envbuilder version**, the **same architecture** (amd64), and the
**same repo content** (Dockerfile, `devcontainer.json`, feature versions).
When you bump the envbuilder version, update it in both
`.github/workflows/cache-devcontainer.yml` and the Coder template.
```

- [ ] **Step 2: Validate the section is present**

Run:
```bash
for needle in "prebuilt image cache on GHCR" "ghcr.io/bmorton/devcontainer-cache" \
  "ENVBUILDER_CACHE_REPO" "ENVBUILDER_GET_CACHED_IMAGE" "ghcr.io/coder/envbuilder:1.3.0"; do
  grep -qF "$needle" README.md && echo "ok: $needle" || { echo "MISSING: $needle" >&2; exit 1; }
done
```
Expected: an `ok:` line for each needle, exit 0.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "Document GHCR devcontainer cache and Coder consumption

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Notes for the executor

- Tasks are independent and can be implemented in order 1 → 2 → 3. Task 2 and 3
  reference artifacts from Task 1 but only by name/contract, so each is testable
  on its own.
- Docker is **not** available in the authoring devcontainer; that's expected.
  `scripts/build-cache.sh` is validated via its `DRY_RUN` path. The real
  build+push only runs on the `namespace-profile-devcontainer` CI runner.
- Final end-to-end validation (out of band, after merge): trigger the workflow
  via `workflow_dispatch`, confirm the image appears at
  `ghcr.io/bmorton/devcontainer-cache`, make the package public, then confirm a
  fast cached boot in Coder.
