## Context

Currently, when a workspace starts, it uses `sleep infinity` as the container entrypoint. The existing codebase has agent functionality in `agent/server.go` that provides SSH server capabilities.

This design changes the container entrypoint from `sleep infinity` to the agent binary. The agent runs as PID 0 (init process) and:
1. Provides SSH server for remote shell access
2. Provides gRPC service for command dispatch and status reporting
3. SSH and gRPC share the same port using protocol detection (mux)
4. Forks child processes to execute user commands
5. Handles zombie process reaping (init process responsibility)

## Goals / Non-Goals

**Goals:**
- Add CLI flags (`--agent` / `--no-agent`) to `up` and `start` commands to control agent injection
- Display SSH connection details after workspace starts with agent enabled
- Expose agent status in workspace metadata and `codepod list` output
- Make agent port configurable via `CODEPOD_AGENT_PORT` environment variable
- Pass environment variables to agent process in container
- Run agent as container's PID 0 (init process)
- Agent forks child processes for user tasks
- Provide gRPC service for command dispatch and status reporting
- SSH and gRPC share the same port using protocol detection/mux

**Non-Goals:**
- Adding authentication beyond password (SSH key-based auth can be added later)
- Implementing agent health monitoring via separate endpoint (status via gRPC)
- Multi-agent support (single agent per workspace only)

## Decisions

1. **Agent enabled by default**: Agent injection will be enabled by default when starting a workspace. Users can disable it with `--no-agent` flag.

2. **Fixed port range**: Agent will use ports in range 22001-22010 to avoid conflicts. Each workspace gets a unique port.

3. **Configuration via environment variables**: Agent settings controlled via environment variables:
   - `CODEPOD_AGENT_PORT`: SSH/gRPC port for agent (default: 22001)
   - `CODEPOD_AGENT_PASSWORD`: Password for SSH access (default: "codepod")

4. **CLI flags for injection control**: Use `--agent` / `--no-agent` flags to enable/disable injection, while environment variables control the agent's behavior.

5. **Graceful failure**: If agent injection fails, the workspace still starts. A warning is shown but it's not a fatal error.

6. **Agent as PID 0**: Instead of using `sleep infinity` as container entrypoint, run agent binary as PID 0 (init process). Agent runs SSH server and forks child processes for user tasks.

7. **Agent forks child processes**: Agent runs as init process (PID 0), spawning child processes to handle user commands and managing process lifecycle (reaping zombie processes).

8. **SSH + gRPC port sharing**: SSH and gRPC share the same port using protocol detection:
   - First bytes of connection determine protocol (gRPC uses HTTP/2 magic bytes)
   - SSH connections handled by SSH server
   - gRPC connections handled by gRPC service
   - Simplifies firewall rules and port management

## Risks / Trade-offs

- **[Risk]** Container restart loses agent: The agent process is not persisted across container restarts.
  - **[Mitigation]** Agent is re-injected automatically on each `Start` call (already implemented)

- **[Risk]** Port conflicts if multiple workspaces try to use same port.
  - **[Mitigation]** Use port pool allocation to assign unique ports

- **[Risk]** Agent binary not found on host.
  - **[Mitigation]** Show clear error message with instructions to build the agent

## Migration Plan

1. Add `--agent` flag to `codepod up` and `codepod start` commands
2. Update workspace types to include agent configuration
3. Display SSH connection info after successful workspace start
4. Update `codepod list` to show agent status
5. Add environment variable support for agent configuration (CODEPOD_AGENT_PORT, CODEPOD_AGENT_PASSWORD)
6. Pass environment variables to agent process in container
7. Change container entrypoint from `sleep infinity` to agent binary
8. Agent runs as PID 0 (init process) in container
9. Agent implements init process functionality: fork child processes, reap zombies
10. Implement gRPC service for command dispatch and status reporting
11. Implement SSH+gRPC port multiplexing on same port

## Open Questions

- Should agent password be configurable via environment variable for automation?
- Should we add a `codepod agent status <workspace>` command to check agent health?
