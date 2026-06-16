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
- GitHub Copilot CLI
- Anthropic Claude Code
- opencode
- `@openprose/prose-cli` (npm)
- Namespace CLI (`nsc`)

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
