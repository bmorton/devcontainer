FROM mcr.microsoft.com/devcontainers/typescript-node:0-20

RUN apt-get update -y && \
      apt-get install -y \
        build-essential \
        git \
        postgresql-client \
        curl

# Install kubectl (latest stable version)
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Anthropic Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Install opencode-ai (AI coding agent for terminal)
RUN npm install -g opencode-ai

# Authentication setup for Claude Code CLI:
# Set your Anthropic API key as an environment variable
# Example: export ANTHROPIC_API_KEY=your_api_key_here
# Or create a .env file in your project with: ANTHROPIC_API_KEY=your_api_key_here
# Get your API key from: https://console.anthropic.com/account/keys
