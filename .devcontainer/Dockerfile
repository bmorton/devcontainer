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

# Install Playwright with the Chromium browser and its required OS dependencies.
# Browsers are stored in a shared, system-wide location (PLAYWRIGHT_BROWSERS_PATH)
# so both the `node` and `root` users can run Playwright without re-downloading.
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
RUN npm install -g playwright \
    && playwright install --with-deps chromium \
    && chmod -R a+rx /ms-playwright

# Make `/workspaces` owned by the non-root `node` user so cloning repos into it
# works. The base image creates `/workspaces` as root, and envbuilder (used by
# Coder) does not chown the workspace to the target user, so `node` otherwise
# can't create directories there:
#   git clone ... -> fatal: could not create work tree dir: Permission denied
RUN mkdir -p /workspaces && chown node:node /workspaces

# Tell dev container features where the target user's home directory is.
# The devcontainer CLI passes `_REMOTE_USER_HOME`/`_CONTAINER_USER_HOME` to every
# feature's install.sh, but envbuilder (used by Coder) only sets `_REMOTE_USER`
# and `_CONTAINER_USER` -- it leaves the *_HOME variables empty. Features that
# rely on them then break: e.g. the claude-code feature runs
# `cp "$_REMOTE_USER_HOME/.local/bin/claude" /usr/local/bin/claude`, which under
# envbuilder expands to `cp /.local/bin/claude ...` and fails the whole build.
# Supplying the values here (as build ARGs, so they don't leak into the running
# container's environment) keeps feature installs working on the envbuilder path.
# ARGs declared here remain in scope for the feature install steps envbuilder
# appends after this Dockerfile, but reset for the devcontainer CLI (which sets
# these variables itself), so this is a no-op there.
ARG _REMOTE_USER_HOME=/home/node
ARG _CONTAINER_USER_HOME=/home/node

# Default to the non-root `node` user. envbuilder (used by Coder) picks its target
# user from the last `USER` directive when `containerUser`/`remoteUser` are not
# applied from image metadata, so make the intended user explicit here to avoid
# dropping into a root shell.
USER node
