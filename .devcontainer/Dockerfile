FROM mcr.microsoft.com/devcontainers/typescript-node:0-20

RUN apt-get update -y && \
      apt-get install -y \
        build-essential \
        git \
        postgresql-client

# Install Anthropic Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Authentication setup for Claude Code CLI:
# Set your Anthropic API key as an environment variable
# Example: export ANTHROPIC_API_KEY=your_api_key_here
# Or create a .env file in your project with: ANTHROPIC_API_KEY=your_api_key_here
# Get your API key from: https://console.anthropic.com/account/keys
