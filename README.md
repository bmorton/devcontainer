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
- gnome-keyring / libsecret (Linux Secret Service for VS Code SecretStorage)

## Secret storage

A per-container Linux Secret Service (`gnome-keyring`) is started on
container start/attach so VS Code's `SecretStorage` / `keytar` /
`libsecret` consumers (including GitHub Copilot) store credentials in a
local encrypted vault instead of VS Code's plaintext fallback. See
`.devcontainer/keyring-init.sh` and `.devcontainer/keyring-bootstrap.py`.


## VSCode extensions

- eamodio.gitlens
- sourcegraph.cody-ai
- shopify.ruby-lsp
- aki77.rails-db-schema
- golang.go
- ms-kubernetes-tools.vscode-kubernetes-tools
