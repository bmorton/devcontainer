# devcontainer

My personal devcontainer for VS Code and Coder.

## How it is put together

- **Base image:** `mcr.microsoft.com/devcontainers/typescript-node:26-trixie`
  (Node 26 LTS on Debian 13).
- **No devcontainer features.** Everything is installed by
  [`.devcontainer/Dockerfile`](.devcontainer/Dockerfile). Coder's envbuilder
  prebuilt-image cache probe cannot reproduce feature layers
  ([coder/terraform-provider-envbuilder#68](https://github.com/coder/terraform-provider-envbuilder/issues/68)),
  so a feature-based image always rebuilds on workspace start; a pure Dockerfile
  lets workspaces boot straight from the cached image.
- **Tool versions live in [`.devcontainer/mise.toml`](.devcontainer/mise.toml)**
  and are installed system-wide by [mise](https://mise.jdx.dev) during the image
  build. Every version is pinned; the weekly
  [`Bump mise tools`](.github/workflows/bump-mise-tools.yml) workflow opens a PR
  with newer pins. Dependabot covers the base image and GitHub Actions.

### Tools managed by mise

| Area | Tools |
| --- | --- |
| Languages | Go, Ruby, Rust (rustup), uv |
| Go tooling | golangci-lint, staticcheck, gopls, dlv (delve), govulncheck |
| Kubernetes | kubectl, helm, minikube |
| Cloud / platform | GitHub CLI (`gh`), Azure CLI (`az`) |
| Coding agents | Claude Code (`claude`), GitHub Copilot CLI (`copilot`), opencode, pi |
| Terminal | tmux (official static build, newer than Debian's package) |

The exact versions are in `mise.toml`. Tools are reached through mise shims
(`/usr/local/share/mise/shims`, already on `PATH` for every shell, including
non-interactive ones used by VS Code, Coder and CI), so nothing needs
`mise activate`. Project checkouts can carry their own `mise.toml` to override
a version for that project; `mise use` / `mise install` work as usual.

### Installed directly by the Dockerfile

- APT: build-essential, git, postgresql-client, curl, htop, vim, openssh-server,
  and the headers native Ruby gems need.
- sshd on port 2222, started by `postStartCommand`.
- Namespace CLI (`nsc`).
- Playwright with Chromium and its OS dependencies (`PLAYWRIGHT_BROWSERS_PATH=/ms-playwright`).
- `/workspaces` is owned by `node` so `git clone` works there on Coder.
- tmux configuration for `node` and `root`.

### Shared tool homes

`GOPATH=/go`, `CARGO_HOME=/usr/local/cargo` and `RUSTUP_HOME=/usr/local/rustup`
are system-wide and owned by `node`, so `go install` and `cargo install` work
without sudo and survive a persisted home directory.

## VS Code extensions

- eamodio.gitlens
- sourcegraph.cody-ai
- shopify.ruby-lsp
- aki77.rails-db-schema
- golang.go
- ms-kubernetes-tools.vscode-kubernetes-tools

## tmux, iTerm2 and Claude Code

[`.devcontainer/tmux.conf`](.devcontainer/tmux.conf) keeps the `C-a` prefix and
custom key bindings, and is configured so terminal-aware tools behave inside
tmux the way they do outside it:

| Setting | Why |
| --- | --- |
| `default-terminal tmux-256color`, `terminal-features xterm*:RGB,...` | True color and modern terminfo inside panes. |
| `allow-passthrough all` | Claude Code's desktop notifications, tab progress bar and inline images reach iTerm2/Kitty/Ghostty, even from a window that is not visible. |
| `focus-events on` | Applications learn when their pane gains/loses focus. |
| `extended-keys on` + `extkeys` feature | Shift+Enter inserts a newline in Claude Code instead of submitting. |
| `set-clipboard on` | OSC 52 copies (Claude Code `/copy`) land on the local clipboard. |
| `escape-time 10` | Esc is instant in Claude Code and vim. |

On the iTerm2 side, once: Settings → Profiles → Terminal → enable
"Notification Center Alerts" and, under "Filter Alerts", "Send escape
sequence-generated alerts"; Settings → General → Selection → "Applications in
terminal may access clipboard" (running `/terminal-setup` in Claude Code does
this one for you, even from inside tmux). Restart Claude Code after changing
tmux settings; it only inspects `$TMUX` at startup.

iTerm2's **Session Status** indicator for Claude Code (OSC 7501) is not relayed
by a normal `tmux attach`. To get it inside tmux, attach with iTerm2's tmux
integration instead: `tmux -CC attach` (iTerm2 3.7.5+ shows Claude Code sessions
running in tmux-integration panes).

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

## Faster Coder startup: prebuilt image cache on GHCR

Coder builds this dev container with [envbuilder](https://github.com/coder/envbuilder),
which otherwise rebuilds the whole image on every workspace start.

The [`Cache Devcontainer`](.github/workflows/cache-devcontainer.yml) workflow
runs the same envbuilder and pushes a fully cached image to
`ghcr.io/bmorton/devcontainer-cache` on every push to `main`, weekly, and on
manual dispatch. Coder then reuses the cache instead of rebuilding. Because the
image is now feature-free, the `envbuilder_cached_image` probe can match it and
boot directly from the prebuilt image.

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
**same repo content** (Dockerfile, `devcontainer.json`, `mise.toml`, `tmux.conf`).
When you bump the envbuilder version, update it in both
`.github/workflows/cache-devcontainer.yml` and the Coder template.

## Verifying a build

`scripts/verify-devcontainer.sh` runs inside the built container (CI uses it as
envbuilder's init script) and asserts: user is `node`, tmux starts with the
shipped config, the lifecycle completed, `/workspaces` is writable, and every
mise-pinned tool is installed and runnable from a plain `PATH` lookup.
`scripts/test-envbuilder.sh` reproduces the Coder build locally with Docker.
