FROM mcr.microsoft.com/devcontainers/typescript-node:24-bookworm

# Wire up the tmux configuration for the default container user (node)
COPY tmux.conf /home/node/.tmux.conf
RUN chown node:node /home/node/.tmux.conf

# Also wire up tmux configuration for root (used when connecting via `coder ssh`)
COPY tmux.conf /root/.tmux.conf
RUN chown root:root /root/.tmux.conf
