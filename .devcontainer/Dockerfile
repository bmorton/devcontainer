# syntax=docker/dockerfile:1
#
# Dockerfile-only devcontainer (no devcontainer `features:`).
#
# Why: Coder builds workspaces with envbuilder, and envbuilder's prebuilt-image
# cache probe (the `envbuilder_cached_image` Terraform resource) cannot reproduce
# devcontainer *feature* layers -- features install from `/.envbuilder/features/...`
# during the build but from `/tmp/...` during the probe, so the probe always misses
# and the workspace rebuilds. See coder/terraform-provider-envbuilder#68.
#
# Folding every tool into plain Dockerfile RUN layers lets the probe reproduce the
# build byte-for-byte, so a cached workspace can boot directly from the prebuilt
# image (no build, no layer extraction). The toolset mirrors the previous
# feature-based devcontainer.json.
FROM mcr.microsoft.com/devcontainers/typescript-node:24-bookworm

# All install steps run as root; the final USER directive drops back to `node`.
USER root
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive

# ---------------------------------------------------------------------------
# APT packages: the requested toolbox plus Ruby build dependencies and sshd.
# ---------------------------------------------------------------------------
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential git postgresql-client curl htop tmux vim \
        ca-certificates gnupg lsb-release wget unzip \
        openssh-server \
        autoconf bison libssl-dev libyaml-dev libreadline-dev zlib1g-dev \
        libncurses-dev libffi-dev libgdbm-dev libdb-dev uuid-dev \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------
# tmux configuration for both `node` (default) and `root` (used by `coder ssh`).
# ---------------------------------------------------------------------------
COPY tmux.conf /home/node/.tmux.conf
COPY tmux.conf /root/.tmux.conf
RUN chown node:node /home/node/.tmux.conf && chown root:root /root/.tmux.conf

# ---------------------------------------------------------------------------
# Go (was ghcr.io/devcontainers/features/go, version 1.26.3). System-wide.
# ---------------------------------------------------------------------------
ENV GO_VERSION=1.26.3
ENV GOROOT=/usr/local/go GOPATH=/go
ENV PATH=${GOROOT}/bin:${GOPATH}/bin:${PATH}
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm /tmp/go.tgz \
    && mkdir -p "${GOPATH}/bin" "${GOPATH}/src" "${GOPATH}/pkg" \
    && chmod -R a+rwX "${GOPATH}"

# ---------------------------------------------------------------------------
# Rust (was ghcr.io/devcontainers/features/rust). System-wide via rustup.
# ---------------------------------------------------------------------------
ENV RUSTUP_HOME=/usr/local/rustup CARGO_HOME=/usr/local/cargo
ENV PATH=${CARGO_HOME}/bin:${PATH}
RUN curl --proto '=https' --tlsv1.2 -fsSL https://sh.rustup.rs \
        | sh -s -- -y --no-modify-path --profile minimal --default-toolchain stable \
    && rustup component add clippy rustfmt \
    && chmod -R a+rwX "${RUSTUP_HOME}" "${CARGO_HOME}"

# ---------------------------------------------------------------------------
# Ruby (was ghcr.io/devcontainers-extra/features/ruby-asdf, version 4.0.5).
# Built with ruby-build and installed system-wide to /usr/local/ruby.
# ---------------------------------------------------------------------------
ENV RUBY_VERSION=4.0.5
ENV PATH=/usr/local/ruby/bin:${PATH}
RUN git clone --depth 1 https://github.com/rbenv/ruby-build.git /tmp/ruby-build \
    && PREFIX=/usr/local /tmp/ruby-build/install.sh \
    && ruby-build "${RUBY_VERSION}" /usr/local/ruby \
    && rm -rf /tmp/ruby-build \
    && /usr/local/ruby/bin/gem install bundler --no-document \
    && chmod -R a+rwX /usr/local/ruby/lib/ruby/gems

# ---------------------------------------------------------------------------
# kubectl, helm, minikube (was ghcr.io/devcontainers/features/kubectl-helm-minikube).
# ---------------------------------------------------------------------------
RUN KUBECTL_VERSION="$(curl -fsSL https://dl.k8s.io/release/stable.txt)" \
    && curl -fsSL "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" -o /usr/local/bin/kubectl \
    && chmod 0755 /usr/local/bin/kubectl \
    && curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash \
    && curl -fsSL https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 -o /usr/local/bin/minikube \
    && chmod 0755 /usr/local/bin/minikube

# ---------------------------------------------------------------------------
# GitHub CLI (was ghcr.io/devcontainers/features/github-cli).
# ---------------------------------------------------------------------------
RUN mkdir -p -m 755 /etc/apt/keyrings \
    && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------
# Azure CLI (was ghcr.io/devcontainers/features/azure-cli, latest).
# ---------------------------------------------------------------------------
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------
# sshd (was ghcr.io/devcontainers/features/sshd). Listens on 2222.
# ---------------------------------------------------------------------------
RUN mkdir -p /var/run/sshd \
    && ssh-keygen -A \
    && sed -i 's/^#\?Port .*/Port 2222/' /etc/ssh/sshd_config \
    && sed -i 's/^#\?PasswordAuthentication .*/PasswordAuthentication yes/' /etc/ssh/sshd_config

# ---------------------------------------------------------------------------
# Namespace (nsc) CLI - https://namespace.so/docs/reference/cli/installation
# ---------------------------------------------------------------------------
RUN curl -fsSL https://get.namespace.so/cloud/install.sh | NS_INSTALL_DIR=/usr/local/bin sh

# ---------------------------------------------------------------------------
# Playwright + Chromium in a shared, system-wide browser location.
# ---------------------------------------------------------------------------
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
RUN npm install -g playwright \
    && playwright install --with-deps chromium \
    && chmod -R a+rx /ms-playwright

# ---------------------------------------------------------------------------
# Global npm CLIs (replacing the claude-code, opencode, copilot-cli and
# npm-package features). Installed system-wide via the base image's Node.
#   - @anthropic-ai/claude-code  -> `claude`
#   - @github/copilot            -> `copilot`
#   - opencode-ai                -> `opencode`
#   - @openprose/prose-cli       -> prose CLI
# ---------------------------------------------------------------------------
RUN npm install -g \
        @anthropic-ai/claude-code \
        @github/copilot \
        opencode-ai \
        @openprose/prose-cli

# Default to the non-root `node` user. envbuilder (used by Coder) picks its
# target user from the last `USER` directive when `containerUser`/`remoteUser`
# are not applied from image metadata, so make the intended user explicit.
USER node
