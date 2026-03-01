# Data Model: WSL容器开发环境

**Feature**: 001-wsl-container-dev
**Date**: 2026-03-01

## 实体定义

### 1. Config (系统配置)

```yaml
version: "1.0"
wsl:
  distribution: "Ubuntu-22.04"  # 默认WSL发行版
  docker_host: "tcp://localhost:2375"  # Docker端点
general:
  default_ide: "vscode"  # 默认IDE
  ssh_port: 2222  # 默认SSH端口
```

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| Version | string | 是 | 配置版本 |
| Wsl.Distribution | string | 是 | WSL发行版名称 |
| Wsl.DockerHost | string | 是 | Docker端点 |
| General.DefaultIDE | string | 是 | 默认IDE |
| General.SSHPort | int | 是 | 默认SSH端口 |

### 2. Workspace

```yaml
name: "myproject"
created_at: "2026-03-01T10:00:00Z"
updated_at: "2026-03-01T10:00:00Z"
state: "running"  # created | running | stopped | error
repository:
  url: "https://github.com/user/repo"
  branch: "main"
ide:
  type: "vscode"  # vscode | jetbrains | other
  settings: {}  # IDE特定设置
container:
  image: "myproject-dev:latest"
  name: "myproject-dev-container"
domain: "myproject.codepod"
ssh:
  config_path: "~/.ssh/config"
  key_path: "~/.codepod/keys/myproject"
agent:
  port: 2222
  status: "running"  # running | stopped | error
```

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| Name | string | 是 | Workspace名称（唯一） |
| CreatedAt | timestamp | 是 | 创建时间 |
| UpdatedAt | timestamp | 是 | 更新时间 |
| State | enum | 是 | 状态 |
| Repository.URL | string | 是 | Git仓库URL |
| Repository.Branch | string | 是 | Git分支 |
| IDE.Type | string | 是 | IDE类型 |
| IDE.Settings | object | 否 | IDE设置 |
| Container.Image | string | 是 | 容器镜像 |
| Container.Name | string | 是 | 容器名称 |
| Domain | string | 是 | 访问域名 |
| SSH.ConfigPath | string | 是 | SSH配置路径 |
| SSH.KeyPath | string | 是 | SSH密钥路径 |
| Agent.Port | int | 是 | Agent端口 |
| Agent.Status | enum | 是 | Agent状态 |

### 3. 环境状态

```
未创建 (created)
    ↓ 创建
运行中 (running)
    ↓ 停止
已停止 (stopped)
    ↓ 启动
运行中 (running)

运行中 (running)
    ↓ 错误
错误 (error)
    ↓ 修复
运行中 (running)
```

| 状态 | 描述 |
|------|------|
| created | 已创建但未启动 |
| running | 运行中 |
| stopped | 已停止 |
| error | 发生错误 |

## 验证规则

### Config验证
- Wsl.Distribution: 必须为已安装的WSL发行版
- Wsl.DockerHost: 必须为有效的Docker端点格式
- General.SSHPort: 必须在1024-65535范围内

### Workspace验证
- Name: 必须符合DNS域名标签（字母数字+连字符，最大63字符）
- Repository.URL: 必须为有效的Git URL格式
- Domain: 必须以 .codepod 结尾

## 关系图

```
Config (全局配置)
  ↑
Workspace (工作空间)
  ├── Repository (代码仓)
  ├── IDE (IDE配置)
  ├── Container (容器)
  ├── Domain (域名)
  ├── SSH (SSH配置)
  └── Agent (Agent)
```

## 存储

- 配置文件: `~/.codepod/config.yaml`
- Workspace数据: `~/.codepod/workspaces/{name}.yaml`
- SSH密钥: `~/.codepod/keys/{name}/`
