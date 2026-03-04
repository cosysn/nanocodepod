## Context

Currently, codepod CLI runs on Windows and directly interacts with WSL for Docker operations. This creates complexity:
- Path conversion issues between Windows and WSL
- Complex file copy operations
- Direct Docker command execution via WSL

The new architecture introduces a server component that runs inside WSL:
- **Deployment time**: CLI interacts with WSL to deploy the server
- **Runtime**: CLI communicates with server via gRPC - no more complex WSL interaction!

## Goals / Non-Goals

**Goals:**
- Run codepod server inside WSL to access Docker directly
- Server manages container lifecycle (create, start, stop, remove)
- Server copies agent binary into each container during creation
- Windows client communicates with server for workspace operations
- Support multiple providers (e.g., different WSL distributions)

**Non-Goals:**
- Cloud deployment or multi-server architecture
- Authentication/authorization between client and server
- Container orchestration across multiple servers

## Decisions

### Decision 1: Server runs inside WSL

**Choice**: The server component runs inside WSL where Docker is available.

**Rationale**: This solves the Docker access issue on Windows. The server has direct access to Docker in WSL.

### Decision 2: gRPC for communication

**Choice**: Use gRPC for client-server communication.

**Rationale**: gRPC provides efficient binary serialization, works well with Protobuf, and supports streaming.

### Decision 3: Agent deployed by server

**Choice**: Server copies agent binary into containers during creation.

**Rationale**: Server has access to both the agent binary (copied from Windows) and Docker, so it's the natural place to handle agent deployment.

### Decision 4: Provider Interface

**Choice**: Provider manages the environment itself (not individual containers).

**Rationale**: Provider controls the WSL environment:

```go
interface Provider {
    Init() error           // Initialize environment (install Docker, etc.)
    Start() error          // Start environment (e.g., wsl -d <distro>)
    Stop() error           // Stop environment (e.g., wsl --shutdown)
    Status() (Status, error) // Get environment status (running/stopped)
    Command(cmd) (output, error)  // Execute command/script in environment
}
```

- `Start`: 开启 WSL（如 `wsl -d <distro>`）
- `Stop`: 停止 WSL（如 `wsl --shutdown` 或 `wsl -t <distro>`）
- `Status`: 查看 WSL 状态（运行中/已停止）
- `Init`: 初始化环境（安装 Docker 等）
- `Command`: 将脚本传入 WSL 执行（类似 VSCode 的 WSL 管道）

### Decision 5: WSL Provider Implementation

**Choice**: WSL provider uses shell script injection via WSL pipe.

**Rationale**: Similar to VSCode's WSL injection:
- Prepare shell script on Windows
- Pass script to WSL via `wsl.exe bash -c`
- Execute in WSL
- Server runs in WSL, handles Command execution via gRPC

### Decision 6: Provider-specific config file on Windows

**Choice**: Each provider has its own config file stored on Windows.

**Rationale**: Provider config is managed by CLI on Windows:
- Location: `~/.codepod/providers/<name>/config.yaml`
- Contains: wsl_distribution, data_dir, server_port, etc.

## Risks / Trade-offs

- **Risk**: Network latency between Windows client and WSL server
  - **Mitigation**: gRPC is efficient; for most operations this should be acceptable

- **Risk**: Server dependency on WSL path access
  - **Mitigation**: Use the wsl$ path mechanism for file transfer

- **Risk**: Complexity of managing two components
  - **Mitigation**: Server can run as background process; client handles discovery

## Deployment Model

### Release Package
- Build creates a compressed archive (zip/tar.gz)
- Contains both x86 and arm64 binaries:
  - `server-linux-amd64`
  - `server-linux-arm64`
  - `agent-linux-amd64`
  - `agent-linux-arm64`

### Deployment Directory
- `~/.codepod-server/bin/<commitid>/`
- CLI deploys correct binary based on WSL architecture

## Migration Plan

1. Build script creates release package with all platform binaries
2. CLI "deploy" command extracts correct binary to `~/.codepod-server/bin/<commitid>/`
3. Server started from deployed location
4. Client connects to server via gRPC

## Open Questions

- How to version/track deployed commitid?
- Server port configuration?
