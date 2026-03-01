# Implementation Plan: WSL容器开发环境

**Branch**: `001-wsl-container-dev` | **Date**: 2026-03-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-wsl-container-dev/spec.md`

## Summary

在Windows上基于WSL构建一个容器开发环境，类似devpod、daytona、coder但更简单。核心功能包括：
1. TUI界面提供友好的进度显示和错误提示
2. Workspace管理（关联代码仓、IDE，开发容器，域名，SSH配置）
3. Devcontainer支持（集成devcon工具）
4. Agent注入（提供SSH Server、Git Forward、监控）
5. CLI配置管理（WSL发行版、Docker端点等）
6. 持久化存储（WSL目录映射到容器）
7. 端口分配管理（避免端口冲突）
8. IDE自动启动（VS Code、JetBrains系列）
9. 幂等支持（重入安全）

主要技术方案：Go CLI + WSL2 + Docker + devcon

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: bubbletea (TUI框架), spf13/cobra (CLI框架), docker/client, golang.org/x/crypto/ssh
**Storage**: 本地文件系统 (YAML配置文件存放在 ~/.codepod/)
**Testing**: Go testing, integration tests
**Target Platform**: Windows + WSL2
**Project Type**: CLI工具 (命令行工具)
**Performance Goals**:
- 环境创建时间 < 60秒
- 容器启动时间 < 30秒
- TUI响应时间 < 100ms
**Constraints**:
- 需要WSL2和Docker Desktop
- 配置文件存放在 ~/.codepod/
- devcon工具路径: /home/ubuntu/devcon/
- Agent提供SSH Server（不依赖sshd）
**Scale/Scope**:
- 单用户本地开发环境
- 10-50个并发容器
- 100 LOC/命令

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

根据Constitution原则，本项目需要满足：

| 原则 | 状态 | 说明 |
|------|------|------|
| Test-First | ✅ PASS | 测试先行，Red-Green-Refactor |
| All Tests Must Pass | ✅ PASS | 提交前所有测试必须通过 |
| Test Coverage >= 80% | ✅ PASS | 单元测试覆盖率要求 |
| Test Case Documentation | ✅ PASS | 测试用例存放在 docs/cases/ |
| Bug Tracking | ✅ PASS | Bug记录存放在 docs/bugs.md |

**结论**: Constitution Check PASS，无需复杂度过高 justification。

## Project Structure

### Documentation (this feature)

```
specs/001-wsl-container-dev/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-commands.md # CLI命令契约
└── tasks.md            # Phase 2 output
```

### Source Code

```text
codepod/                 # CLI主程序
├── cmd/
│   ├── up.go            # 创建并启动Workspace (推荐)
│   ├── create.go        # 仅创建Workspace
│   ├── list.go          # 列出Workspace
│   ├── start.go         # 启动Workspace
│   ├── stop.go          # 停止Workspace
│   ├── delete.go        # 删除Workspace
│   ├── connect.go       # 连接Workspace (自动打开IDE)
│   ├── config.go        # 配置管理
│   └── root.go          # Root命令
├── internal/
│   ├── config/          # 配置管理模块
│   ├── workspace/        # Workspace管理
│   ├── wsl/             # WSL交互
│   ├── docker/          # Docker操作
│   ├── agent/           # Agent注入
│   ├── devcon/          # devcon集成
│   ├── tui/             # TUI界面
│   ├── ide/             # IDE启动
│   ├── storage/         # 持久化存储
│   ├── port/            # 端口分配
│   └── types/           # 类型定义
├── tests/
├── docs/
│   ├── cases/           # 测试用例
│   └── bugs.md          # Bug记录
└── main.go

agent/                   # Agent程序
├── main.go
├── ssh/                 # SSH Server
├── git/                 # Git Forward
└── monitor/            # 监控
```

**Structure Decision**:
- codepod: CLI主程序，使用Go + bubbletea + cobra
- agent: 轻量级Agent程序，提供SSH Server、Git Forward、监控
- 配置文件: ~/.codepod/config.yaml
- Workspace数据: ~/.codepod/workspaces/
- 工具目录: ~/.codepod/tools/ (devcon)

## Complexity Tracking

无需复杂度追踪，Constitution Check全部通过。

---

## Phase 0: Research (Research.md)

### 技术选型总结

| 组件 | 技术选择 | 理由 |
|------|----------|------|
| CLI框架 | spf13/cobra | 成熟稳定，自动生成帮助 |
| TUI框架 | bubbletea | 声明式架构，易测试 |
| Docker客户端 | docker/client | 官方SDK |
| SSH Server | golang.org/x/crypto | 纯Go实现，无需sshd |
| 配置文件 | gopkg.in/yaml.v3 | 易读易写 |
| IDE启动 | exec.Command | 调用系统命令打开IDE |

---

## Phase 1: Design

### 已生成文档

- [x] research.md - 技术选型研究
- [x] data-model.md - 数据模型设计
- [x] quickstart.md - 快速开始指南
- [x] contracts/cli-commands.md - CLI命令契约

### CLI命令

| 命令 | 描述 |
|------|------|
| `codepod up <name>` | 创建并启动Workspace，自动打开IDE |
| `codepod create <name>` | 仅创建Workspace |
| `codepod list` | 列出Workspace |
| `codepod start <name>` | 启动Workspace |
| `codepod stop <name>` | 停止Workspace |
| `codepod delete <name>` | 删除Workspace |
| `codepod connect <name>` | 连接Workspace，自动打开IDE |
| `codepod config` | 配置管理 |

### IDE支持

| IDE | 启动方式 |
|-----|----------|
| VS Code | `code --remote ssh-remote+...` |
| JetBrains | `jetbrains-goland://...` |

---

## Next Steps

1. 运行 `/speckit.tasks` 生成任务列表
2. 开始实现
