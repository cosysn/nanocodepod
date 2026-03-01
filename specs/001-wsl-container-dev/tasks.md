# Tasks: WSL容器开发环境

**Feature**: 001-wsl-container-dev
**Generated**: 2026-03-01
**Plan**: [plan.md](./plan.md)
**Spec**: [spec.md](./spec.md)

## Task Summary

| 统计 | 数量 |
|------|------|
| 总任务数 | 45 |
| 用户故事数 | 11 (P1: 10, P2: 1) |
| 可并行任务 | 12 |

---

## Phase 1: Setup (项目初始化)

- [X] T001 Initialize Go project with go.mod
- [X] T002 [P] Setup project directory structure per plan.md
- [X] T003 Install dependencies: cobra, bubbletea, docker/client, yaml

---

## Phase 2: Foundational (共享基础设施)

- [X] T004 Create internal/types package with Workspace, Config, Agent structs
- [X] T005 [P] Create internal/types/workspace.go with Workspace model
- [X] T006 [P] Create internal/types/config.go with Config model
- [X] T007 Create internal/config package for config load/save
- [X] T008 Create internal/wsl package for WSL command execution
- [X] T009 Create internal/docker package for Docker operations
- [X] T010 Create internal/storage package for persistent storage
- [X] T011 Create internal/port package for port allocation

---

## Phase 3: User Story 2 - 配置存储 (P1)

**Goal**: All config files stored in ~/.codepod/

**Independent Test**: Check ~/.codepod/ directory created with proper structure

- [X] T012 [US2] Implement config initialization in internal/config/init.go
- [X] T013 [US2] Create ~/.codepod directory structure on init
- [X] T014 [US2] Implement workspace config save/load in internal/config/workspace.go
- [ ] T015 [US2] Write test cases for config storage in tests/config/

---

## Phase 4: User Story 3 - CLI配置管理 (P1)

**Goal**: CLI commands to manage configuration

**Independent Test**: Run `codepod config set/get/list` and verify

- [X] T016 [US3] Create cmd/config.go with cobra subcommands
- [X] T017 [US3] Implement config set command in cmd/config_set.go
- [X] T018 [US3] Implement config get command in cmd/config_get.go
- [X] T019 [US3] Implement config list command in cmd/config_list.go

---

## Phase 5: User Story 1 - TUI界面 (P1)

**Goal**: Friendly TUI with progress and error messages

**Independent Test**: Run commands and see TUI output

- [X] T020 [US1] Create internal/tui package with bubbletea components
- [X] T021 [US1] Implement progress display for long operations
- [X] T022 [US1] Implement error display with solutions
- [ ] T023 [US1] Implement help message display

---

## Phase 6: User Story 9 - Devcontainer支持 (P1)

**Goal**: Integrate devcon tool and build container images

**Independent Test**: Run build and verify image created

- [X] T024 [US9] Create internal/devcon package for devcon integration
- [X] T025 [US9] Implement devcon injection to WSL
- [X] T026 [US9] Implement image build using devcon
- [X] T027 [US9] Stream build output to TUI

---

## Phase 7: User Story 10 - 端口分配管理 (P1)

**Goal**: Auto-assign unique ports to avoid conflicts

**Independent Test**: Create multiple workspaces, verify no port conflicts

- [X] T028 [US10] Implement port pool management in internal/port/pool.go
- [X] T029 [US10] Implement port allocation in internal/port/allocate.go
- [X] T030 [US10] Implement port release in internal/port/release.go

---

## Phase 8: User Story 11 - 持久化存储 (P1)

**Goal**: Persistent storage for workspace code

**Independent Test**: Delete and recreate container, verify code persists

- [X] T031 [US11] Implement storage directory creation in internal/storage/create.go
- [X] T032 [US11] Implement volume mapping for Docker in internal/storage/mount.go
- [X] T033 [US11] Implement storage cleanup on delete in internal/storage/cleanup.go

---

## Phase 9: User Story 8 - Agent注入 (P1)

**Goal**: Inject agent with SSH, Git Forward, monitor

**Independent Test**: SSH connect to container, run git commands, check monitor

- [X] T034 [US8] Create agent/ program structure
- [X] T035 [US8] Implement SSH server in agent/ssh/server.go
- [X] T036 [US8] Implement Git forward in agent/git/forward.go
- [X] T037 [US8] Implement monitor in agent/monitor/stats.go
- [X] T038 [US8] Implement agent injection to container

---

## Phase 10: User Story 4 - 创建开发环境 (P1)

**Goal**: Create workspace with container

**Independent Test**: Run `codepod up` and verify container running

- [X] T039 [US4] Create cmd/up.go with create and start logic
- [X] T040 [US4] Create cmd/create.go with idempotent create
- [X] T041 [US4] Implement workspace creation flow in internal/workspace/create.go

---

## Phase 11: User Story 5 - 管理开发环境 (P1)

**Goal**: Start, stop, delete, list workspaces

**Independent Test**: Run list/start/stop/delete commands

- [X] T042 [US5] Create cmd/list.go for workspace listing
- [X] T043 [US5] Create cmd/start.go for workspace start
- [X] T044 [US5] Create cmd/stop.go for workspace stop
- [X] T045 [US5] Create cmd/delete.go for workspace deletion

---

## Phase 12: User Story 6 - 连接开发环境 (P1)

**Goal**: Connect to workspace with IDE auto-launch

**Independent Test**: Run `codepod connect` and verify IDE opens

- [X] T046 [US6] Create cmd/connect.go
- [X] T047 [US6] Implement IDE launch in internal/ide/launch.go
- [X] T048 [US6] Support VS Code remote connect
- [X] T049 [US6] Support JetBrains IDE connect

---

## Phase 13: User Story 7 - 管理Workspace (P1)

**Goal**: Workspace with repo, IDE, container, domain, SSH config

**Independent Test**: Create workspace with all associations

- [X] T050 [US7] Implement workspace domain config in internal/workspace/domain.go
- [X] T051 [US7] Implement SSH config generation in internal/workspace/ssh_config.go
- [X] T052 [US7] Implement /etc/hosts update for domain

---

## Phase 14: User Story 13 - 自定义开发环境 (P2)

**Goal**: Custom language, resources, Dockerfile

**Independent Test**: Create workspace with custom config

- [ ] T053 [US13] Implement resource limits in internal/workspace/resources.go
- [ ] T054 [US13] Support custom Dockerfile in devcon build

---

## Phase 15: Polish (收尾)

- [ ] T055 Add integration tests for full workflow
- [X] T056 Verify idempotency for all commands
- [ ] T057 Add test cases to docs/cases/
- [ ] T058 Update quickstart.md with final commands
- [X] T059 Build and test codepod binary

---

## Dependencies

```
Phase 1 (Setup)
  │
  └─► Phase 2 (Foundational)
        │
        ├─► Phase 3 (US2: 配置存储)
        │     └─► Phase 4 (US3: CLI配置管理)
        │
        ├─► Phase 5 (US1: TUI界面)
        │     └─► Phase 6 (US9: Devcontainer)
        │
        ├─► Phase 7 (US10: 端口分配)
        ├─► Phase 8 (US11: 持久化存储)
        ├─► Phase 9 (US8: Agent注入)
        │
        └─► Phase 10 (US4: 创建开发环境)
              └─► Phase 11 (US5: 管理开发环境)
                    └─► Phase 12 (US6: 连接开发环境)
                          └─► Phase 13 (US7: 管理Workspace)
                                └─► Phase 14 (US13: 自定义)
                                      └─► Phase 15 (Polish)
```

## Parallel Opportunities

- T002, T004, T005, T006 can run in parallel (setup tasks)
- T028, T029, T030 can run in parallel (port management)
- T031, T032, T033 can run in parallel (storage management)
- T034, T035, T036, T037 can run in parallel (agent components)
- T046, T047, T048, T049 can run in parallel (IDE connect)

## MVP Scope

For MVP, implement:

- Phase 1: Setup
- Phase 2: Foundational (T004-T011)
- Phase 3: US2 配置存储
- Phase 4: US3 CLI配置管理
- Phase 6: US9 Devcontainer支持 (core functionality)
- Phase 7: US10 端口分配
- Phase 8: US11 持久化存储
- Phase 9: US8 Agent注入
- Phase 10: US4 创建开发环境

This provides a working `codepod up` command that creates and starts a workspace with agent.
