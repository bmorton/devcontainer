FROM mcr.microsoft.com/devcontainers/typescript-node:24-bookworm

# Install a Linux Secret Service backend so VS Code's SecretStorage /
# keytar / libsecret consumers (incl. GitHub Copilot) get a local
# encrypted credential vault instead of the plaintext fallback store.
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        gnome-keyring \
        libsecret-1-0 \
        libsecret-tools \
        dbus \
        python3-jeepney \
    && rm -rf /var/lib/apt/lists/*

# Wire up the tmux configuration for the default container user (node)
COPY tmux.conf /home/node/.tmux.conf
RUN chown node:node /home/node/.tmux.conf

# Also wire up tmux configuration for root (used when connecting via `coder ssh`)
COPY tmux.conf /root/.tmux.conf
RUN chown root:root /root/.tmux.conf
