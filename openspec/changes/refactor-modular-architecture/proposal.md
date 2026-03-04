# Proposal: Refactor to Modular Architecture

## Summary

Refactor codepod from a monolithic structure into separate CLI, Server, and Agent modules to improve maintainability and follow the principle of separation of concerns.

## Problem Statement

Currently, the codepod codebase has several issues:

1. **Mixed concerns**: CLI, agent, and various backend functionalities are in a single module
2. **Heavy CLI**: CLI contains too much logic including workspace management, Docker operations, port allocation, storage management
3. **Hard to test**: Monolithic structure makes unit testing difficult
4. **Poor separation**: Agent code is mixed with CLI code

## Proposed Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Windows/macOS Host                        │
│                                                                  │
│   CLI (轻量)                                                    │
│   - 命令解析 (Cobra)                                            │
│   - TUI 显示 (进度、列表、状态)                                   │
│   - Provider: 环境检测、Server 发现与路由                         │
│   - IDE 启动 (VS Code, JetBrains)                               │
│                                                                  │
└─────────────────────────┬────────────────────────────────────────┘
                         │ HTTP/gRPC
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
┌─────────────────┐ ┌───────────┐ ┌─────────────────┐
│  WSL Server     │ │ Linux     │ │ Remote Server   │
│ (Windows)       │ │ Server    │ │ (macOS/Remote) │
└─────────────────┘ └───────────┘ └─────────────────┘
         │               │               │
         └───────────────┼───────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Server (Backend)                            │
│                                                                  │
│   - Workspace 管理 (SQLite)                                      │
│   - Docker 操作                                                  │
│   - Port 分配                                                   │
│   - Storage 管理                                                │
│   - Devcontainer                                               │
│   - Config 管理                                                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                         ▲
                         │
┌─────────────────────────┴─────────────────────────────────────┐
│                      Container                                  │
│                                                                  │
│   Agent (SSH + gRPC)                                           │
│   - 向 Server 报心跳                                            │
│   - 接收 Server 指令                                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Module Responsibilities

| Module | Location | Responsibilities |
|--------|----------|-----------------|
| CLI | `cli/` | Command parsing, TUI display, Provider (environment detection, server routing), IDE launch |
| Server | `server/` | Workspace management, Docker operations, Port allocation, Storage management, Devcontainer, Config management |
| Agent | `agent/` | SSH+gRPC server in container, heartbeat reporting to Server |

### Provider Mechanism

- **Purpose**: Shield environment differences (Windows/WSL/macOS/Linux)
- **Server Discovery**: Login to node via Provider's Command interface, execute commands to collect Server info (port, status)
- **Mapping**: Each Provider corresponds to one Server instance

### Agent-Server Communication

```
Server                    Container
  │                          │
  │  1. Inject agent        │
  │ ──────────────────────▶ │
  │  2. Start agent        │
  │     (pass env vars:     │
  │      SERVER_URL)        │
  │                          │
  │  3. Heartbeat           │
  │ ◀────────────────────── │
  │                          │
  │  4. gRPC communication  │
  │ ◀──────────────────────▶ │
```

### Data Storage

- **SQLite**: Persist workspace state
- Each workspace record contains: name, status, container ID, port, provider info, etc.

## Benefits

1. **Clear separation**: Each module has focused responsibilities
2. **Easy testing**: Modules can be tested independently
3. **Flexible deployment**: CLI is lightweight, Server can run on any node
4. **Provider abstraction**: Easy to add new environment support
5. **Better maintainability**: Smaller, focused codebases

## Migration Plan

1. Extract CLI code from `codepod/cmd` to `cli/`
2. Create `agent/` module from `codepod/internal/agent`
3. Expand `server/` to handle all backend logic
4. Add SQLite storage to Server
5. Implement Provider mechanism in CLI
6. Update `go.work` for multi-module coordination
7. Migrate existing workspace data

## Risks

- Breaking changes for existing users
- Need to maintain backward compatibility for config files
- Complex migration process for existing workspaces
