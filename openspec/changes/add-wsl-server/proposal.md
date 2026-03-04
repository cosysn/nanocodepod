## Why

The codebase already has a working prototype with CLI, agent, and WSL integration. However, the architecture has issues:
- CLI directly interacts with WSL (complex, error-prone)
- WSL interaction through Go's exec is cumbersome
- CLI carries too much weight with WSL-specific logic
- No standardized provider system for different environments

This refactoring adds a WSL server to handle WSL-specific operations, making CLI simpler and more extensible.

## What Changes

**Refactoring (existing code to be modified):**
- Refactor CLI to use server instead of direct WSL commands
- Refactor existing agent (already runs in containers)
- Refactor WSL interaction to go through provider interface

**New components:**
- Add WSL server that runs inside WSL and exposes HTTP API
- Add provider interface with init, command, create, delete, start, stop, status methods
- Add Go workspace to manage CLI, server, agent as separate modules
- Add multi-platform build system (Linux/macOS, x86/ARM)

## Existing Code (to be refactored)

**CLI** (`codepod/cmd/`):
- Commands: init, up, start, stop, delete, list, connect
- Uses Cobra framework
- Direct WSL and Docker interaction

**Agent** (`codepod/internal/agent/`):
- SSH + gRPC server running in containers
- Already handles command execution

**WSL** (`codepod/internal/wsl/`):
- WSL distribution management
- Docker access detection

**Workspace Manager** (`codepod/internal/workspace/`):
- Create/start/stop/delete workspaces
- Agent injection

## Capabilities

### New Capabilities
- `wsl-server`: A background server running inside WSL that handles CLI requests for WSL operations
- `wsl-cli-protocol`: The communication protocol between CLI and WSL server (HTTP)
- `agent`: An agent that runs inside development containers, injected by the server (already exists, refactor)
- `build-package`: Build system that produces multi-platform archive with CLI, agent, and server binaries
- `provider`: Configuration system for connecting to different environments (multiple WSL distributions, local, etc.)

### Modified Capabilities
- `cli`: Refactor to use provider interface and server communication
- `wsl`: Refactor to use provider interface

## Impact

- CLI refactored to use HTTP server instead of direct WSL exec
- Agent refactored (already exists, minor improvements)
- WSL code refactored into provider interface
- Build system added for multi-platform releases
