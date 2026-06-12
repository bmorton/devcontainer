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

### Setting a master password

By default the keyring is created with an empty master password, so the
encrypted file is trivial to decrypt from a filesystem snapshot. To
raise the bar, set `KEYRING_PASSWORD`; the init script feeds it to both
the initial collection bootstrap and every subsequent unlock.

In **Codespaces**, add `KEYRING_PASSWORD` as a
[user secret](https://docs.github.com/en/codespaces/managing-your-codespaces/managing-your-account-specific-secrets-for-github-codespaces)
and it is injected automatically.

For a **local Dev Container** (no Codespaces), forward the value from
your host shell through `remoteEnv` in `.devcontainer/devcontainer.json`:

```jsonc
"remoteEnv": {
  "XDG_RUNTIME_DIR": "/tmp/runtime-1000",
  "DBUS_SESSION_BUS_ADDRESS": "unix:path=/tmp/runtime-1000/bus",
  "KEYRING_PASSWORD": "${localEnv:KEYRING_PASSWORD}"
}
```

Then export the password in the shell that launches VS Code before
opening the container (or add it to your shell profile):

```bash
export KEYRING_PASSWORD='your-strong-password'
```

To rotate the password, delete `~/.local/share/keyrings/` inside the
container and rebuild — the bootstrap recreates the vault with the new
value.



## VSCode extensions

- eamodio.gitlens
- sourcegraph.cody-ai
- shopify.ruby-lsp
- aki77.rails-db-schema
- golang.go
- ms-kubernetes-tools.vscode-kubernetes-tools
