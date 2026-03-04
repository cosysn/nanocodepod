## Why

On Windows with Docker installed inside WSL, Windows applications cannot directly access the Docker daemon because Docker runs within the WSL distribution, not on the Windows host. The current code assumes Docker is either directly accessible on Windows (via Docker Desktop with TCP) or running natively on Linux/WSL, but doesn't handle the scenario where Docker is exclusively available inside WSL and needs to be accessed from Windows code.

## What Changes

- Add platform detection for Windows with Docker-in-WSL
- Create a WSL-aware Docker client that executes Docker commands inside WSL when Docker is not directly accessible from Windows
- Add automatic fallback logic: try Windows-native Docker first, then fall back to WSL-based Docker
- Modify Docker client initialization to detect the appropriate execution context

## Capabilities

### New Capabilities

- `wsl-docker-access`: Enables the Docker client to detect and use Docker when it's only available inside a WSL distribution, allowing codepod to work correctly on Windows where Docker runs in WSL

### Modified Capabilities

- None - this is a new capability

## Impact

- Affected code: `codepod/internal/docker/client.go`, `codepod/internal/wsl/platform.go`
- New dependencies: WSL command execution for Docker operations
- The change is backward compatible - existing Windows with Docker Desktop and Linux deployments continue to work
