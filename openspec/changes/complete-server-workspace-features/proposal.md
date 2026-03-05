## Why

The WSL server implementation currently lacks key workspace features compared to the standalone codepod CLI. Users expect full functionality (Git clone, local directory, devcontainer, agent injection, IDE connection) when using the server-based architecture. Without these features, the server mode is incomplete and users must fall back to the CLI for common workflows.

## What Changes

- Add Git repository clone support to workspace creation
- Add local directory mounting/copying support
- Add devcontainer.json parsing and image building
- Add agent binary injection into containers
- Add IDE auto-launch via protocol handlers
- Add explicit Linux provider support with --provider flag

## Capabilities

### New Capabilities

- **workspace-git-clone**: Clone Git repositories during workspace creation
- **workspace-local-dir**: Support local directory as workspace source
- **workspace-devcontainer**: Parse .devcontainer.json and build custom images
- **workspace-agent-inject**: Inject codepod-agent binary into running containers
- **workspace-ide-connect**: Launch IDE via vscode:// or jetbrains:// protocols
- **linux-provider**: Explicit Linux provider with --provider flag for provider selection

### Modified Capabilities

- **workspace-create**: Expand to support git, local-dir, devcontainer, and agent options

## Impact

- Server API: Add new endpoints or extend existing ones for git/devcontainer/agent options
- CLI: Update `workspace create` command to pass new options
- Dependencies: May need git binary in server environment
