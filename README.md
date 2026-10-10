# devcontainer

My personal devcontainer for VSCode.

## Base image

- `mcr.microsoft.com/devcontainers/typescript-node:24-bookworm`

## Devcontainer features

- Go
- Ruby (via asdf)
- Rust
- sshd
- kubectl / helm / minikube
- GitHub CLI (`gh`)
- Azure CLI (`az`, latest)
- GitHub Copilot CLI
- Anthropic Claude Code
- opencode
- `@openprose/prose-cli` (npm)
- Namespace CLI (`nsc`)
- Playwright with the Chromium browser (and required OS dependencies)
- pi coding agent with i-have-adhd and a lean global workflow

## APT packages

- build-essential
- git
- postgresql-client
- curl
- htop
- tmux
- vim

## VSCode extensions

- eamodio.gitlens
- sourcegraph.cody-ai
- shopify.ruby-lsp
- aki77.rails-db-schema
- golang.go
- ms-kubernetes-tools.vscode-kubernetes-tools

## GitHub MCP server with the Copilot CLI

The GitHub Copilot CLI ships with GitHub's MCP server enabled by default, so no
extra MCP configuration is required to talk to GitHub from `copilot`. The CLI (and
its bundled GitHub MCP server) authenticate with the same GitHub credentials.

To wire everything up inside the container:

1. Authenticate the GitHub CLI: `gh auth login` (or set `GH_TOKEN` / `GITHUB_TOKEN`).
2. Launch the Copilot CLI: `copilot`. If you are not already authenticated, run the
   `/login` slash command, or export a token before launching:

   ```bash
   export GH_TOKEN="$(gh auth token)"
   copilot
   ```

3. Confirm the GitHub MCP server is available from within `copilot` with the `/mcp`
   slash command, then ask Copilot to interact with your repositories, issues, and
   pull requests.

## pi coding agent

The image installs `@earendil-works/pi-coding-agent@1.1.0` globally using
`npm install --ignore-scripts`. Both `node` and `root` get their own
`~/.pi/agent/` with:

- Short personal preferences from [AGENTS.md](.devcontainer/pi/AGENTS.md).
- The native [i-have-adhd](https://github.com/ayghri/i-have-adhd) package,
  pinned to commit `723af7d9afaf43eb871dbcce6129e2bf80de90d5`, and its always-on marker.
- Four editable skills: spec-first, plan-slices, verify-done, and debug-loop.
- Three prompt templates: `/spec <feature>`, `/plan <spec section>`, and
  `/slice <plan filename or number> <slice number>`.

These are global resources, so they work in any repository without copying a
skeleton. No superpowers package is installed. The skills create project specs
and plans only when requested; they do not impose the approvals project's stack.
Verification uses `mise run check` only when the project defines it, otherwise
the project's documented checks.

### First use

1. Rebuild the container, then launch `pi` as your normal workspace user.
2. Run `/login` and choose GitHub Copilot if using your Copilot subscription.
3. Use `/model`, then Ctrl+S to save your default model.
4. Start with `/spec <feature>`, review the spec, then `/plan <section>` and
   implement approved slices with `/slice <plan> <n>`.

Authentication and model choices are deliberately not baked into the image.
Persist your user's `~/.pi` using your workspace platform's home persistence or a
private volume if they should survive container replacement. A volume mounted
over this directory hides the image defaults, so seed an empty volume from the
image first. Do not commit authentication files or share this volume between users.
Node and root have separate configuration and login state.

Use `/i-have-adhd off` or `/i-have-adhd on` per session. Remove
`~/.pi/agent/.i-have-adhd-always` to disable the default for new sessions.
Edit the bundled resources under `.devcontainer/pi/` and rebuild to change image
defaults, or edit your own `~/.pi/agent/` and use `/reload` for session resources.
Changes to image defaults do not update an existing persistent volume.

### Optional project-local pi-squad

Install squad only in the target project, not globally in this image:

```bash
cd /path/to/project
pi install -l npm:pi-squad
```

Review project-local packages before granting trust with `/trust`, then restart
pi and use `/squad defaults` and `/squad agents`. Commit `.pi/settings.json`
and any project role overrides in `.pi/squad/agents/`; add `.pi/npm/` and
`.pi/git/` to that project's `.gitignore`. Keep stack-specific instructions in
the project's `AGENTS.md` and squad overrides, not the global preferences.

Use squad for milestones after approving a spec and plan. Start with concurrency
2, make shared contracts dependency edges, and avoid inheriting the entire
session unless needed. Inspect the diff and run the project checks yourself
before recording `squad_review`; use `forkFromTask` for rework. Debugging and
small changes stay solo. Squad state lives separately in `~/.pi/squad/`, so
persist that directory too if you need to resume runs.

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

There are two ways to consume the cache, both using **the same pinned envbuilder
version** as CI (`ghcr.io/coder/envbuilder:1.3.0`):

**1. Reuse cached layers (simplest).** Set this on the workspace's envbuilder:

| Environment variable | Value |
| --- | --- |
| `ENVBUILDER_CACHE_REPO` | `ghcr.io/bmorton/devcontainer-cache` |

envbuilder finds each prebuilt layer already in the registry and pulls it instead
of rebuilding from scratch.

**2. Boot directly from the prebuilt image (fastest).** In the Coder template, use
the [`envbuilder_cached_image`](https://registry.terraform.io/providers/coder/envbuilder/latest/docs)
resource from the `coder/envbuilder` provider, pointed at the same cache repo, and
run the workspace container from its resolved image:

```hcl
resource "envbuilder_cached_image" "cached" {
  count         = data.coder_workspace.me.start_count
  builder_image = "ghcr.io/coder/envbuilder:1.3.0"
  git_url       = local.repo_url
  cache_repo    = "ghcr.io/bmorton/devcontainer-cache"
}

# Then run the workspace container from envbuilder_cached_image.cached[0].image
# (with envbuilder_cached_image.cached[0].env), falling back to a normal
# envbuilder build when the cache is empty.
```

Under the hood this resource probes the cache with envbuilder's
`ENVBUILDER_GET_CACHED_IMAGE` dry-run and, on a hit, lets the workspace start from
the prebuilt image without building. On a cache miss it falls back to a normal
build.

The cache repo is public, so no pull credentials are needed.

### Why the version must match

Cache hits require identical build inputs and tooling between CI and Coder:
the **same envbuilder version**, the **same architecture** (amd64), and the
**same repo content** (Dockerfile, `devcontainer.json`, feature versions).
When you bump the envbuilder version, update it in both
`.github/workflows/cache-devcontainer.yml` and the Coder template.
