## Why

Currently, when a workspace is launched, it uses `sleep infinity` as the container entrypoint, providing no way to interact with the container remotely. By running the agent as the container's PID 0 (init process), we enable SSH access for remote development, allowing Claude Code to work directly within the dev container with full process management capabilities.

## What Changes

- Use agent as container's entrypoint instead of `sleep infinity`
- Agent runs as PID 0 (init process) in the container
- Agent provides SSH server for remote shell access
- Agent provides gRPC service for command dispatch and status reporting
- SSH and gRPC share the same port using protocol detection/mux
- Agent forks child processes to execute user commands
- Agent handles zombie process reaping (init process responsibility)
- Support environment variables to configure agent (port, password)
- Add CLI flags to enable/disable agent per workspace

## Capabilities

### New Capabilities

- `agent-init`: Run agent as container init process (PID 0)
- `ssh-access`: Provide SSH access to the development container via agent
- `grpc-api`: Provide gRPC service for command dispatch and status reporting (shares port with SSH)

### Modified Capabilities

- None - this is a new capability

## Impact

- **Workspace Package**: Change container entrypoint to use agent binary
- **Agent Package**: Run as PID 0, implement init process functionality, multiplex SSH and gRPC on same port
- **CLI**: Add `--agent` flag to `up` and `start` commands to enable/disable agent
- **Environment Variables**: Support `CODEPOD_AGENT_PORT` and `CODEPOD_AGENT_PASSWORD`
