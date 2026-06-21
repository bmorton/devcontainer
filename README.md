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
