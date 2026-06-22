# Pre-cache the devcontainer image on GHCR via GitHub Actions

## Problem

Coder builds this devcontainer with **envbuilder** (not the `@devcontainers/cli`).
Every workspace start rebuilds the image from scratch: the Dockerfile plus all the
heavy devcontainer features (Go, Ruby via asdf, Rust, kubectl/helm/minikube,
Azure CLI, Playwright + Chromium, etc.). This makes workspace startup slow.

envbuilder supports caching built layers in a container registry. If we build and
push a fully cached image to GHCR from CI, Coder's envbuilder can pull cached
layers (or boot directly from the prebuilt image) instead of rebuilding, cutting
startup time dramatically (the envbuilder docs cite ~36m → ~40s, ~98%, for a
comparably complex image).

## Goal

A GitHub Actions workflow that builds this devcontainer with envbuilder and
publishes a cached image to a **public** GHCR repo, plus the Coder-side
configuration and documentation needed to consume it.

## Approach

Use envbuilder's native cache mechanism (Approach A, chosen over layer-cache-only
and plain `docker build`):

- CI runs the **same** envbuilder used by Coder, with
  `ENVBUILDER_CACHE_REPO` + `ENVBUILDER_PUSH_IMAGE=1`, pushing every layer **and**
  a complete prebuilt image to GHCR.
- Coder sets the same `ENVBUILDER_CACHE_REPO`. It can either pull cached layers
  (set `ENVBUILDER_CACHE_REPO` only) or boot directly from the prebuilt image
  (set `ENVBUILDER_GET_CACHED_IMAGE=true`).

Rejected alternatives:

- **Layer-cache only** (no `PUSH_IMAGE`): strict subset of A; can't boot directly
  from a single image.
- **Plain `docker build` + push base image**: only caches the Dockerfile, which is
  the cheap part. The expensive work is the devcontainer features envbuilder
  installs *after* the Dockerfile, which this would not cache.

## Components

### 1. Workflow: `.github/workflows/cache-devcontainer.yml`

- **Triggers:**
  - `push` to `main` (refresh when the devcontainer config changes)
  - `schedule` weekly cron (pick up base-image / feature security updates)
  - `workflow_dispatch` (manual trigger)
- **Permissions:** `contents: read`, `packages: write`.
- **Runner:** `namespace-profile-devcontainer` (same runner as the existing
  `build-containers.yml` jobs; has Docker available).
- **Env (single source of truth):** `ENVBUILDER_IMAGE` pinned to a specific
  envbuilder release tag (not `:latest`), and `CACHE_REPO=ghcr.io/bmorton/devcontainer-cache`.
- **Steps:**
  1. `actions/checkout`.
  2. Build a base64-encoded Docker `config.json` authenticating to `ghcr.io`
     with `${{ github.actor }}` + the workflow `GITHUB_TOKEN`, exported as
     `ENVBUILDER_DOCKER_CONFIG_BASE64` for the script.
  3. Run `scripts/build-cache.sh`.

### 2. Script: `scripts/build-cache.sh`

Mirrors `scripts/test-envbuilder.sh` (repo mounted into the envbuilder container,
`.devcontainer/devcontainer.json` auto-discovered), but configured to publish:

- `ENVBUILDER_CACHE_REPO` (required; from `CACHE_REPO`, e.g.
  `ghcr.io/bmorton/devcontainer-cache`)
- `ENVBUILDER_PUSH_IMAGE=1`
- `ENVBUILDER_EXIT_ON_PUSH_FAILURE=1` (fail loudly if the push fails)
- `ENVBUILDER_CACHE_TTL_DAYS=30` (envbuilder default is 7; a longer TTL avoids
  layers expiring between weekly runs)
- `ENVBUILDER_DOCKER_CONFIG_BASE64` (registry credentials, passed in by the
  workflow)
- `ENVBUILDER_INIT_SCRIPT` pointed at `scripts/verify-devcontainer.sh` so the
  existing startup checks (user is `node`, tmux works, lifecycle sentinel) run as
  part of the cache job and a broken build / scheduled base-image regression
  fails the workflow.
- Honors an `ENVBUILDER_IMAGE` env override (same convention as
  `test-envbuilder.sh`) so the workflow controls the pinned version.

The script validates that `CACHE_REPO` and `ENVBUILDER_DOCKER_CONFIG_BASE64` are
set and that Docker is available before running.

### 3. Coder-side wiring + README documentation

Add a README section documenting how Coder consumes the cache:

- Set the workspace template's envbuilder configuration:
  - `ENVBUILDER_CACHE_REPO=ghcr.io/bmorton/devcontainer-cache`
  - the **same pinned envbuilder version** as CI
  - optionally `ENVBUILDER_GET_CACHED_IMAGE=true` to boot directly from the
    prebuilt image (fastest path)
- Public repo ⇒ no pull credentials required on the Coder side.
- **One-time manual step:** after the first successful push, set the
  `devcontainer-cache` GHCR package visibility to **Public** (GHCR packages are
  created private by default).

## Correctness constraints

Cache hits require the build inputs and tooling to match between CI and Coder:

- **Identical envbuilder version** (hence the pinned tag, used in both places).
- **Same architecture** (amd64 on both the Namespace runner and Coder workspaces).
- **Same repo content** — Dockerfile, `devcontainer.json`, and feature versions.
  Layer hashes are computed from these; any difference busts the cache.

## Verification

- The cache job runs `scripts/verify-devcontainer.sh` as the envbuilder init
  script, so the run is only green if the built container passes the existing
  startup checks.
- Manual validation: trigger via `workflow_dispatch`, confirm the image appears at
  `ghcr.io/bmorton/devcontainer-cache`, then confirm a fast cached boot in Coder.

## Out of scope

- Changing the existing PR verification workflow (`build-containers.yml`).
- Multi-architecture cache images (amd64 only).
- Automating GHCR package visibility (one-time manual step).
