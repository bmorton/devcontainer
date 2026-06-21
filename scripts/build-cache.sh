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
