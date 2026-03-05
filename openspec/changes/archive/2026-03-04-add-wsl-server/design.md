## Context

**Existing codebase (prototype):**
- CLI: Cobra-based commands (init, up, start, stop, delete, list, connect)
- Agent: SSH+gRPC server running in containers (codepod/internal/agent/)
- WSL: Distribution management and Docker detection (codepod/internal/wsl/)
- Workspace: Manager for create/start/stop/delete workspaces

**Current issues:**
- CLI directly executes WSL commands using Go's exec package
- Complex path handling between Windows and WSL
- Error handling across the Windows/WSL boundary is fragile
- CLI is bloated with WSL-specific logic
- No standardized provider interface

**Goal:** Refactor to add WSL server and provider interface while keeping existing functionality.

## Goals / Non-Goals

**Goals:**
- Simplify CLI by moving WSL complexity to a server component
- Provide a clean HTTP/Unix socket API for CLI-server communication
- Support Docker operations, file operations, and process management
- Keep deployment simple - server starts on-demand during deployment

**Non-Goals:**
- Replace all CLI functionality - only WSL-related operations move to server
- Implement authentication/authorization between CLI and server (local only)
- Handle server persistence across WSL restarts (server is ephemeral)

## Decisions

### Decision: HTTP over Unix Socket
**Choice:** HTTP server inside WSL, CLI connects via Windows-side HTTP client

**Alternatives considered:**
- Unix socket: Cleaner but requires Windows-side Unix socket support
- Named pipes: Windows-native but more complex to implement in Go

**Rationale:** HTTP is well-supported on both sides, easy to debug, and the overhead is negligible for local communication.

### Decision: Different protocols for CLI-Server and Agent-Server
**Choice:** CLI communicates with Server via HTTP; Server communicates with Agent via gRPC

**Alternatives considered:**
- Single protocol: Simpler but agent may need different capabilities
- All HTTP: Simpler but loses streaming benefits of gRPC

**Rationale:**
- CLI-Server: HTTP is simple and sufficient for command/response
- Agent-Server: gRPC provides better streaming and existing agent uses gRPC

### Decision: Server embedded in CLI binary
**Choice:** Server code is part of the same Go module, deployed alongside CLI

**Alternatives considered:**
- Separate repository: More complex to maintain
- Script-based server: Less type-safe, harder to integrate

**Rationale:** Single binary deployment is simpler; server is started via `wsl <command>` during deployment phase.

### Decision: Use Go workspace for multi-component management
**Choice:** Use Go workspace to manage CLI, server, and agent as separate modules in one workspace

**Alternatives considered:**
- Single module: Simpler but less modular
- Separate repositories: More complex to build and release

**Rationale:** Go workspace allows building all components from a single source while keeping codebases separate

### Decision: Start server on-demand
**Choice:** Server starts when first CLI request arrives, stays running

**Alternatives considered:**
- Always-on: Wastes resources when not in use
- Manual start: More user interaction required

**Rationale:** "SSH-style" - server starts on first request, subsequent requests reuse the same server instance.

### Decision: Single archive with platform subdirectories
**Choice:** Archive contains platform-specific subdirectories (linux-x86, linux-arm, macos-x86, macos-arm)

**Alternatives considered:**
- Separate archives per platform: More downloads but smaller individual files
- Flat structure: Simpler but harder to identify platform

**Rationale:** Clear organization, users extract to ~/.codepod-server/bin/<commit-id>/

### Decision: Shared commit-id across all binaries
**Choice:** CLI, agent, and server share the same git commit-id for versioning

**Alternatives considered:**
- Separate versioning: More flexibility but harder to track compatibility
- Timestamp-based: Works but less traceable

**Rationale:** Simplifies deployment - all components from same source are compatible

### Decision: Provider configuration in YAML files
**Choice:** Each provider has its own YAML configuration directory in ~/.codepod/provider/<name>/

**Alternatives considered:**
- Single config file: Simpler but harder to manage multiple environments
- Database-backed: More complex, overkill for local use

**Rationale:** Easy to understand, edit, and share provider configs. Provider code is built into CLI.

### Decision: Provider selection via CLI flag
**Choice:** User specifies provider via --provider flag or environment variable

**Alternatives considered:**
- Interactive selection: More user-friendly but slower for scripts
- Default provider: Simpler but less flexible

**Rationale:** Works well for both interactive and automated usage

### Decision: Provider interface for extensibility
**Choice:** Define a provider interface in Go that can be implemented by different environment providers

**Alternatives considered:**
- Hardcoded providers: Simpler but harder to extend
- Plugin system: More flexible but complex to manage

**Rationale:** Interface-based design allows adding AWS, Linux, macOS providers without modifying core code

### Decision: Env provider for deployment
**Choice:** Each provider implements an env sub-provider responsible for deploying container environment and injecting server/agent

**Alternatives considered:**
- Centralized deployment: Less flexible per-provider
- Manual deployment: More error-prone

**Rationale:** Each environment has different deployment requirements; env provider encapsulates them

### Decision: WSL deployment via shell script pipe
**Choice:** Prepare shell script locally, pipe it to WSL for execution (similar to VSCode WSL injection)

**Alternatives considered:**
- Direct command execution: Less reliable for complex deployments
- SCP then SSH: More complex, similar result

**Rationale:** VSCode uses this approach successfully; it's reliable and handles complex multi-step deployments

### Decision: Agent runs in development containers
**Choice:** Agent is injected into containers created by the server

**Alternatives considered:**
- Server handles all operations: Less flexible
- Separate agent deployment: More complex

**Rationale:** Agent runs inside the developer's container, can execute commands on behalf of the server

### Decision: Deployment downloads from GitHub release
**Choice:** Installation script downloads archive from GitHub release

**Alternatives considered:**
- Bundle with CLI: CLI becomes larger
- Build from source: Takes longer

**Rationale:** Download on-demand keeps CLI small, same archive used for all deployments

## Risks / Trade-offs

- [Risk] Server crash leaves CLI hanging → Mitigation: Timeout with clear error message
- [Risk] Server port conflict → Mitigation: Use random available port, communicated via stdout
- [Risk] WSL distribution not running → Mitigation: CLI starts WSL distribution automatically before connecting

## Migration Plan

1. User adds provider: `codepod provider add <name> --type wsl --wsl-distro <distro>`
2. CLI creates provider config at ~/.codepod/provider/<name>/config.yaml
3. CLI calls provider init to initialize environment
4. CLI calls provider command to pipe installation script to WSL
5. Script downloads archive from GitHub release to ~/.codepod-server/bin/<commit-id>/
6. Script extracts archive and installs CLI, server, agent binaries
7. Script starts codepod-server in background
8. Server creates development containers and injects agent into them
9. CLI connects to server for operations
