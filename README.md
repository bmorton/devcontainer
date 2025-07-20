# devcontainer

This is my personal devcontainer for VSCode.

## Tools Included

### kubectl

This devcontainer includes the latest stable version of kubectl for Kubernetes development and management tasks. kubectl is automatically installed during container build and is available in the PATH.

### Anthropic Claude Code CLI

This devcontainer includes the [Anthropic Claude Code CLI](https://github.com/anthropics/claude-code), an agentic coding tool that helps you code faster through natural language commands.

#### Authentication

To use Claude Code, you need to set up your Anthropic API key:

1. Get your API key from [Anthropic Console](https://console.anthropic.com/account/keys)
2. Set the environment variable:
   ```bash
   export ANTHROPIC_API_KEY=your_api_key_here
   ```
3. Or add it to your shell profile or project's `.env` file

#### Usage

Once authenticated, you can use Claude Code in your terminal:

```bash
# Navigate to your project directory
cd /workspaces/your-project

# Start Claude Code
claude

# Claude Code will help you with:
# - Code explanations and reviews
# - Automated coding tasks
# - Git workflow assistance
# - Natural language to code conversion
```

For more information, see the [official Claude Code documentation](https://docs.anthropic.com/en/docs/claude-code/overview).

### Coder CLI

This devcontainer includes the [Coder CLI](https://coder.com), a command-line interface for Coder, a cloud-based development environment platform that allows you to create and manage remote development environments.

#### Usage

The Coder CLI allows you to:

```bash
# Login to your Coder deployment
coder login <your-coder-url>

# List available workspaces
coder list

# Create a new workspace
coder create

# Connect to a workspace via SSH
coder ssh <workspace-name>

# Start/stop workspaces
coder start <workspace-name>
coder stop <workspace-name>
```

For more information, see the [official Coder documentation](https://coder.com/docs).

### opencode-ai

This devcontainer includes [opencode-ai](https://www.npmjs.com/package/opencode-ai), an AI coding agent built for the terminal with responsive UI and automatic LSP loading capabilities.

#### Features

- Responsive terminal UI
- Automatic LSP (Language Server Protocol) loading
- Support for multiple agents working in parallel
- AI-powered coding assistance

#### Usage

You can use opencode-ai directly from the terminal:

```bash
# Navigate to your project directory
cd /workspaces/your-project

# Start opencode-ai
opencode

# opencode-ai provides:
# - AI-powered code generation and editing
# - Intelligent code suggestions
# - Terminal-based coding assistance
# - Multi-agent support for complex tasks
```
