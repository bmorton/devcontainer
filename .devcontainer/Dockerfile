# syntax=docker/dockerfile:1
#
# Dockerfile-only devcontainer: no devcontainer `features:`.
#
# Coder builds this workspace with envbuilder, and envbuilder's prebuilt-image
# cache probe (the `envbuilder_cached_image` Terraform resource) cannot reproduce
# devcontainer *feature* layers, so a feature-based image always misses the cache
# and rebuilds on every workspace start (coder/terraform-provider-envbuilder#68).
# Everything is therefore installed from plain RUN layers here. Tool versions
# live in mise.toml (installed system-wide with mise); only the few things mise
# cannot manage (apt packages, nsc, Playwright) are installed directly.
#
# The build context is the REPOSITORY ROOT (devcontainer.json: build.context
# "..", CI: `docker build -f .devcontainer/Dockerfile .`) so the in-container
# utilities in utilities/ can be compiled here. COPY paths are repo-relative.
FROM mcr.microsoft.com/devcontainers/typescript-node:26-trixie

USER root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG DEBIAN_FRONTEND=noninteractive

# ---------------------------------------------------------------------------
# APT packages: general toolbox, sshd, and the headers native Ruby gems and
# ruby-build need (libyaml/openssl/readline/ffi/gdbm/zlib).
# tmux is intentionally NOT installed from apt; a current release comes from mise.
# ---------------------------------------------------------------------------
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential git postgresql-client curl htop vim \
        ca-certificates unzip \
        openssh-server \
        autoconf bison libssl-dev libyaml-dev libreadline-dev zlib1g-dev \
        libncurses-dev libffi-dev libgdbm-dev libdb-dev uuid-dev \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------
# sshd (replaces ghcr.io/devcontainers/features/sshd). Listens on 2222 and is
# started by postStartCommand in devcontainer.json.
# ---------------------------------------------------------------------------
RUN mkdir -p /var/run/sshd \
    && ssh-keygen -A \
    && sed -i 's/^#\?Port .*/Port 2222/' /etc/ssh/sshd_config \
    && sed -i 's/^#\?PasswordAuthentication .*/PasswordAuthentication yes/' /etc/ssh/sshd_config

# ---------------------------------------------------------------------------
# /workspaces is created root-owned by the base image and envbuilder does not
# chown it, so `git clone` into it failed for the node user. Hand it over.
# ---------------------------------------------------------------------------
RUN mkdir -p /workspaces && chown node:node /workspaces

# ---------------------------------------------------------------------------
# tmux configuration for both `node` (default) and `root` (used by `coder ssh`).
# ---------------------------------------------------------------------------
COPY .devcontainer/tmux.conf /home/node/.tmux.conf
COPY .devcontainer/tmux.conf /root/.tmux.conf
RUN chown node:node /home/node/.tmux.conf && chown root:root /root/.tmux.conf

# ---------------------------------------------------------------------------
# Namespace (nsc) CLI - https://namespace.so/docs/reference/cli/installation
# Installed system-wide so it is on PATH for both `node` and `root`.
# ---------------------------------------------------------------------------
RUN curl -fsSL https://get.namespace.so/cloud/install.sh | NS_INSTALL_DIR=/usr/local/bin sh

# ---------------------------------------------------------------------------
# Playwright with Chromium and its OS dependencies. Browsers live in a shared,
# system-wide location so both users can run Playwright without re-downloading.
# ---------------------------------------------------------------------------
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
RUN npm install -g playwright \
    && playwright install --with-deps chromium \
    && chmod -R a+rx /ms-playwright \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------
# mise (https://mise.jdx.dev) manages every versioned tool: languages, k8s
# CLIs, gh/az, coding agents, tmux. Pinned binary copied from the official
# image; versions pinned in mise.toml (-> /etc/mise/config.toml).
#
# Layout:
#   - MISE_DATA_DIR=/usr/local/share/mise : installs + shims, shared by all
#     users and outside $HOME so a mounted/persisted home cannot shadow it.
#   - shims dir on PATH via ENV, so tools work in non-interactive shells
#     (VS Code server, Coder agent, CI `devcontainer exec`) without `mise
#     activate`. Interactive shells may still `eval "$(mise activate bash)"`.
#   - GOPATH / CARGO_HOME / RUSTUP_HOME are system-wide and node-owned so
#     `go install` / `cargo install` work and persist in the image.
# ---------------------------------------------------------------------------
COPY --from=ghcr.io/jdx/mise:2026.10.23 /usr/local/bin/mise /usr/local/bin/mise
ENV MISE_DATA_DIR=/usr/local/share/mise \
    MISE_TRUSTED_CONFIG_PATHS=/etc/mise \
    GOPATH=/go \
    RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo
ENV PATH=${MISE_DATA_DIR}/shims:${GOPATH}/bin:${CARGO_HOME}/bin:${PATH}
COPY .devcontainer/mise.toml /etc/mise/config.toml
RUN mkdir -p "${MISE_DATA_DIR}" "${GOPATH}" "${RUSTUP_HOME}" "${CARGO_HOME}" \
    && chown -R node:node "${MISE_DATA_DIR}" "${GOPATH}" "${RUSTUP_HOME}" "${CARGO_HOME}"

# Login shells do not see the ENV PATH above: Debian's /etc/profile resets PATH
# and the base image's /etc/profile.d/00-restore-env.sh only restores the PATH
# that existed when the base image was built. `coder ssh`, VS Code terminals and
# new tmux windows are all login shells, so without this the tools "vanish"
# (CI's init script is non-login and never noticed). Export the tool dirs and
# mise variables from a profile.d snippet and source it from the interactive rc
# files too, exactly as the devcontainer features used to do.
COPY .devcontainer/10-devcontainer-tools.sh /etc/profile.d/10-devcontainer-tools.sh
RUN chmod 0644 /etc/profile.d/10-devcontainer-tools.sh \
    && printf '\n. /etc/profile.d/10-devcontainer-tools.sh\n' >> /etc/bash.bashrc \
    && printf '\n. /etc/profile.d/10-devcontainer-tools.sh\n' >> /etc/zsh/zshrc

# Install as `node` so the shared tool tree is owned by the user who will add
# to it (gems, `go install`, `cargo install`, `mise use`). root can still run
# and modify everything.
USER node
RUN MISE_YES=1 mise install --jobs 8 \
    && mise reshim \
    && missing="$(mise ls --missing)" \
    && if [ -n "$missing" ]; then echo "mise tools missing after install:" "$missing" >&2; exit 1; fi \
    && rm -rf /home/node/.cache/mise

# ---------------------------------------------------------------------------
# In-container utilities (utilities/): Go TUIs such as `bootstrap`, which walks
# through logins and identity on a fresh container. Built with the mise-managed
# Go as `node` so the module cache lands in node-owned /go; the binaries are
# then installed system-wide. See utilities/AGENTS.md for the pattern.
# ---------------------------------------------------------------------------
COPY --chown=node:node utilities /usr/local/src/devcontainer-utilities
RUN cd /usr/local/src/devcontainer-utilities \
    && go build -trimpath -ldflags="-s -w" -o /tmp/devcontainer-utilities-bin/ ./cmd/... \
    && go clean -cache \
    && rm -rf /home/node/.cache/go-build
USER root
RUN install -m 0755 /tmp/devcontainer-utilities-bin/* /usr/local/bin/ \
    && rm -rf /tmp/devcontainer-utilities-bin

# Stay on the non-root `node` user. envbuilder (used by Coder) picks its target
# user from the last `USER` directive when `containerUser`/`remoteUser` are not
# applied from image metadata, so the intended user must be explicit here to
# avoid dropping into a root shell.
USER node
