FROM mcr.microsoft.com/devcontainers/typescript-node:24-bookworm

# Wire up the tmux configuration for the default container user (node)
COPY tmux.conf /home/node/.tmux.conf
RUN chown node:node /home/node/.tmux.conf

# Also wire up tmux configuration for root (used when connecting via `coder ssh`)
COPY tmux.conf /root/.tmux.conf
RUN chown root:root /root/.tmux.conf

# Install the Namespace (nsc) CLI - https://namespace.so/docs/reference/cli/installation
# Install system-wide so it is on PATH for both the `node` and `root` users.
RUN curl -fsSL https://get.namespace.so/cloud/install.sh | NS_INSTALL_DIR=/usr/local/bin sh
