## Why

Currently, codepod CLI on Windows has complex WSL interactions for every operation:
- Path conversion between Windows and WSL
- File copy via wsl$ paths
- Direct Docker command execution through WSL

This is too heavy. By adding a server in WSL:
- **Deploy time**: CLI deploys server to WSL once
- **Runtime**: CLI talks to server via gRPC - simple and clean!

## What Changes

- Use Go workspace to manage multiple components:
  - `cli/`: Windows CLI client
  - `server/`: Server runs in WSL
  - `agent/`: Agent runs in containers
- Build all components into release package
- Package contains both x86 and arm64 binaries
- Deployment directory: `~/.codepod-server/bin/<commitid>/`
- Server runs inside WSL, handles all Docker operations
- Agent binary deployed by server into containers
- CLI becomes lightweight client, talks to server via gRPC
- Provider interface for extensibility (WSL, AWS, Linux, macOS, etc.)
- Provider manages environment (WSL):
  - Init: 初始化环境（安装 Docker 等）
  - Start: 开启 WSL
  - Stop: 停止 WSL
  - Status: 查看 WSL 状态
  - Command: 将脚本传入 WSL 执行（通过 WSL 管道，类似 VSCode）
- WSL provider: shell script injection via WSL pipe
- Each provider has config file on Windows (`~/.codepod/providers/<name>/config.yaml`)
- Provider config: data_dir, wsl_distribution, server_port

## Capabilities

### New Capabilities
- `codepod-server`: Server component that runs in WSL, manages Docker containers
- `container-agent-deployment`: Server copies agent binary into containers during creation
- `server-client-communication`: gRPC communication between Windows client and WSL server

### Modified Capabilities
- None - this is a new capability

## Impact

- New code: `cmd/server/` - server component
- Modified code: `cmd/` - Windows client refactored to communicate with server
- Dependencies: gRPC for client-server communication
- The server runs in WSL and manages all Docker operations
