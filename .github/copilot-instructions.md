# Copilot Coding Agent Instructions

This repository contains a personal devcontainer configuration for VSCode with various development tools pre-installed.

## Repository Structure

- `.devcontainer/` - Contains the Dockerfile and devcontainer.json configuration
- `.github/workflows/` - Contains CI/CD workflows for building containers
- `README.md` - Documentation for the devcontainer and its included tools

## Development Workflow

### Building the Devcontainer

This repository uses Docker to build a development container. There are two ways to build:

1. **Build the Dockerfile only:**
   ```bash
   docker build -t devcontainer-dockerfile .devcontainer/
   ```

2. **Build the full devcontainer (recommended):**
   ```bash
   npm install -g @devcontainers/cli
   devcontainer build --workspace-folder .
   ```

### Testing Changes

- All changes to the devcontainer configuration should be tested by building the container
- Use the GitHub Actions workflow to validate builds on pull requests
- The workflow runs automatically on PRs targeting the `main` branch

## Code Standards

### Dockerfile Changes

- Use multi-line RUN commands with backslashes for readability
- Always update package lists before installing packages (`apt-get update -y`)
- Group related installations together
- Add comments explaining what tools are being installed and why

### JSON Configuration

- Use 2-space indentation for JSON files
- Maintain consistent formatting with existing configuration
- Include comments where helpful (JSON5 format is supported in devcontainer.json)

### Documentation

- Update the README.md when adding new tools or features to the devcontainer
- Include authentication/setup instructions for tools that require API keys or configuration
- Provide usage examples for new tools

## Included Tools

The devcontainer includes:
- **Base**: TypeScript/Node.js 20, PostgreSQL client, build-essential, git, curl
- **Languages**: Go 1.24.5, Ruby 3.4.4, Rust
- **Kubernetes**: kubectl (latest stable)
- **AI Tools**: Anthropic Claude Code CLI, opencode-ai
- **VSCode Extensions**: GitLens, Cody AI, Ruby LSP, Go, Kubernetes Tools

## Important Notes

- Do not remove existing tools or features without good reason
- Maintain compatibility with the base image (mcr.microsoft.com/devcontainers/typescript-node:0-20)
- Keep security in mind - do not add credentials or secrets to the container configuration
- Test that the postCreateCommand works correctly after changes to kubectl installation

## CI/CD

The repository has a GitHub Actions workflow (`.github/workflows/build-containers.yml`) that:
- Builds the Dockerfile separately (20 minute timeout)
- Builds the full devcontainer (30 minute timeout)
- Runs on all pull requests to the main branch

Ensure your changes pass both build steps before submitting.
