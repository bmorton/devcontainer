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
- GitHub Copilot CLI
- Anthropic Claude Code
- opencode
- `@openprose/prose-cli` (npm)

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
