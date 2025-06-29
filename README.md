# devcontainer

This is my personal devcontainer for VSCode.

## Tools Included

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
