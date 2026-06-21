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

# 4. Optional overrides (ENVBUILDER_IMAGE, ENVBUILDER_CACHE_TTL_DAYS) are reflected.
out=$(CACHE_REPO="ghcr.io/bmorton/devcontainer-cache" \
      ENVBUILDER_DOCKER_CONFIG_BASE64="ZHVtbXk=" \
      ENVBUILDER_IMAGE="ghcr.io/coder/envbuilder:9.9.9" \
      ENVBUILDER_CACHE_TTL_DAYS="7" DRY_RUN=1 bash "$target" 2>&1); rc=$?
check "override exits 0" test "$rc" -eq 0
check "ENVBUILDER_IMAGE override reflected" grep -q "envbuilder_image=ghcr.io/coder/envbuilder:9.9.9" <<<"$out"
check "ENVBUILDER_CACHE_TTL_DAYS override reflected" grep -q "cache_ttl_days=7" <<<"$out"

if [ "$failures" -ne 0 ]; then echo "${failures} check(s) failed." >&2; exit 1; fi
echo "All build-cache tests passed."
