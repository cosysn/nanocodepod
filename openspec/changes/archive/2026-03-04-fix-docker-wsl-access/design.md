## Context

Currently, the Docker client in codepod executes Docker CLI commands directly using `exec.Command("docker", ...)`. On Windows with Docker installed inside WSL, the Docker daemon runs inside the WSL distribution and is not directly accessible from Windows applications.

The platform detection in `platform.go` already distinguishes between Windows, WSL, and Linux platforms, and provides a `GetDockerHost()` function that returns TCP or Unix socket addresses. However, this doesn't solve the fundamental issue: on Windows without Docker Desktop (or when Docker Desktop is configured to run in WSL), the `docker` command itself isn't available in Windows' PATH.

The solution requires detecting when Docker must be executed inside WSL and creating a WSL-aware Docker client wrapper.

## Goals / Non-Goals

**Goals:**
- Enable codepod to work on Windows when Docker is installed inside a WSL distribution
- Maintain backward compatibility with existing deployments (Linux, Windows with Docker Desktop)
- Provide automatic detection and fallback logic for Docker access

**Non-Goals:**
- Modify Docker installation or configuration
- Support multiple WSL distributions simultaneously
- Handle Docker-in-WSL2 vs Docker-in-WSL1 differences at a low level (the approach works for both)

## Decisions

### Decision 1: Execute Docker via WSL when not available on Windows

**Choice**: Use WSL command execution to run Docker commands inside the WSL distribution when Docker is not directly accessible on Windows.

**Rationale**: This is the simplest approach that doesn't require users to modify their Docker configuration. The alternative (configuring Docker to expose its API via TCP on Windows) requires user intervention and security considerations.

**Alternatives considered**:
- Configure Docker in WSL to listen on TCP and access from Windows: Requires modifying Docker config, potential security risk
- Use Docker Contexts: More complex to set up and not all users have this configured

### Decision 2: Fallback detection chain

**Choice**: Try Windows-native Docker first, then fall back to WSL-based Docker.

**Rationale**: Most users either have Docker Desktop installed on Windows (works natively) or have Docker in WSL only (needs WSL execution). The fallback chain handles both cases transparently.

**Detection order**:
1. Try `docker` command directly on Windows
2. If that fails, check if WSL is available and has Docker
3. Use WSL-based execution if detected

### Decision 3: Wrap the existing Docker client

**Choice**: Create a WSL-aware Docker client that wraps the existing CLI-based client.

**Rationale**: The existing `Client` struct in `client.go` already implements all necessary Docker operations via CLI. We can add a thin wrapper layer that executes commands via WSL when needed, reusing all the existing logic.

## Risks / Trade-offs

- **Risk**: WSL command execution may be slower than native execution.
  - **Mitigation**: The overhead is minimal for typical Docker operations. Most container operations are I/O-bound anyway.

- **Risk**: Different WSL distributions may have different Docker setups.
  - **Mitigation**: Default to the primary WSL distribution (from `WSL_DISTRO_NAME` env var or "Ubuntu"), with clear error messages if Docker isn't found.

- **Risk**: Path handling between Windows and WSL.
  - **Mitigation**: Volume bindings and file paths used in container operations need to be translated. The WSL integration in `platform.go` already handles file copying; we need to ensure path translation works for Docker volume mounts.

## Migration Plan

1. Add WSL detection to the Docker client factory
2. Create a wrapper that executes Docker commands via WSL when needed
3. Update the Docker client initialization to use the appropriate execution context
4. Existing deployments are unaffected - the change only activates when Docker is detected as unavailable on Windows and available in WSL
