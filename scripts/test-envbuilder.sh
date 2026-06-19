#!/usr/bin/env bash
#
# Build this dev container the same way Coder does (with envbuilder) and run the
# verification checks, so you can reproduce and debug Coder-only startup issues
# (e.g. "drops to root", "tmux won't run", "something isn't completing")
# locally and read the full envbuilder build log.
#
# Coder builds workspaces with envbuilder, NOT the `@devcontainers/cli`. The two
# diverge on user selection, UID/GID remapping and feature install order, so the
# CLI can succeed while Coder fails. This script exercises the envbuilder path.
#
# Usage:
#   scripts/test-envbuilder.sh
#
# Requirements:
#   - docker
#
# Optional environment variables:
#   ENVBUILDER_IMAGE   envbuilder image to use (default: ghcr.io/coder/envbuilder:latest)
#   CACHE_DIR          host directory for the envbuilder layer cache (speeds up reruns)

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
envbuilder_image="${ENVBUILDER_IMAGE:-ghcr.io/coder/envbuilder:latest}"

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker is required but not found on PATH" >&2
  exit 1
fi

# The repo is mounted into the envbuilder workspace. envbuilder discovers
# `.devcontainer/devcontainer.json` automatically. After the build + lifecycle
# commands complete, it runs ENVBUILDER_INIT_SCRIPT as the target user; we point
# that at the verification script (resolved from the mounted workspace).
workspace_folder="/workspaces/devcontainer"
init_script="bash ${workspace_folder}/scripts/verify-devcontainer.sh"

docker_args=(
  run --rm
  -e "ENVBUILDER_WORKSPACE_FOLDER=${workspace_folder}"
  -e "ENVBUILDER_INIT_SCRIPT=${init_script}"
  -v "${repo_root}:${workspace_folder}"
)

# Optionally persist the layer cache between runs for faster iteration.
if [ -n "${CACHE_DIR:-}" ]; then
  mkdir -p "${CACHE_DIR}"
  docker_args+=(-v "${CACHE_DIR}:/cache" -e "ENVBUILDER_LAYER_CACHE_DIR=/cache")
fi

docker_args+=("${envbuilder_image}")

echo "=== Running envbuilder ($envbuilder_image) on $repo_root ==="
echo "docker ${docker_args[*]}"
exec docker "${docker_args[@]}"
