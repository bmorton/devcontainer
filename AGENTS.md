# bmorton/devcontainer — guidance for coding agents

Personal devcontainer image used from VS Code and, mainly, Coder (built with
envbuilder). The README explains the "why"; this file is the operating manual.

## Non-negotiables

- **No devcontainer `features:`.** envbuilder's cache probe cannot reproduce
  feature layers, so features force a full rebuild on every workspace start.
  Everything is installed from `.devcontainer/Dockerfile`.
- **Versioned tools go in `.devcontainer/mise.toml`,** pinned exactly. Check
  `mise registry <name>` first (note: the registry's `nsc` is NATS, not
  Namespace). Only apt packages, nsc and Playwright are installed directly.
- **Nothing important lives under `/home/node` in the image.** Coder may persist
  the home directory, which would shadow it. Tools go to `/usr/local/share/mise`,
  `GOPATH=/go`, `CARGO_HOME`/`RUSTUP_HOME` under `/usr/local`, all node-owned.
- **PATH additions must be made twice:** `ENV PATH` for non-login processes
  and `.devcontainer/10-devcontainer-tools.sh` (profile.d, sourced from
  bash.bashrc and zsh/zshrc) for login shells. Debian's `/etc/profile` resets
  PATH and the base image only restores its own.
- **Keep the tmux key bindings.** Everything else in `.devcontainer/tmux.conf`
  may be modernised; test on both the current tmux and the mise-installed one.

## Validating changes

There is no Docker inside the container. The PR workflow `Build Containers`
is the test: plain `docker build`, `devcontainer up` + exec, and an envbuilder
build that runs `scripts/verify-devcontainer.sh` as the init script. Extend
that script whenever you add something the image must guarantee (a binary, a
PATH entry, a config file). On merge, `Cache Devcontainer` republishes the
prebuilt image; Coder workspaces pick it up on next start.

## Repo map

- `.devcontainer/` — Dockerfile, devcontainer.json (build context is the repo
  root), mise.toml, tmux.conf, profile.d snippet.
- `utilities/` — Go module of in-container TUIs (`bootstrap`); has its own
  AGENTS.md and STYLEGUIDE.md.
- `scripts/` — verify script (runs inside the built container), local
  envbuilder reproduction, cache publishing.
- `.github/workflows/` — PR builds, weekly `mise upgrade --bump` PR, cache publish.

## Habits

- Dependency bumps: mise.toml via the weekly workflow (or `mise upgrade --bump`
  in `.devcontainer/`); base image and actions via Dependabot.
- Prefer small PRs with the CI run linked; write down the lesson in the
  README or this file when a Coder-only failure is found.
