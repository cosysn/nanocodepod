# Quickstart: WSL容器开发环境

**Feature**: 001-wsl-container-dev
**Date**: 2026-03-01

## 前置条件

### 1. 安装WSL2

```powershell
# 以管理员身份打开PowerShell
wsl --install
# 重启电脑
```

### 2. 安装Docker Desktop

1. 下载 [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)
2. 安装并启动
3. 确保启用 "Enable integration with additional distros"

### 3. 下载devcon工具

devcon工具需要放在 `/home/ubuntu/devcon/` 目录下。

## 安装CodePod

```bash
# 克隆项目
git clone <repo-url>
cd codepod

# 构建
go build -o codepod.exe .

# 或者安装到PATH
go install .
```

## 快速开始

### 1. 初始化

```bash
# 首次运行会自动初始化
codepod init
```

### 2. 配置

```bash
# 设置WSL发行版
codepod config set wsl.distribution Ubuntu-22.04

# 设置Docker端点
codepod config set wsl.docker-host tcp://localhost:2375

# 查看配置
codepod config list
```

### 3. 创建Workspace (自动打开IDE)

```bash
# 从Git仓库创建并启动，自动打开IDE (推荐)
codepod up myproject --repo https://github.com/user/repo --ide vscode

# 指定代码目录，自动打开IDE
codepod up myproject --path /path/to/code --ide vscode

# 仅创建，不启动 (类似 devpod create)
codepod create myproject --repo https://github.com/user/repo
```

**说明**: `codepod up` 会自动:
1. 创建Workspace和持久化存储
2. 调用devcon构建镜像
3. 创建并启动容器
4. 注入Agent
5. 自动打开IDE并连接到开发环境

### 4. 管理Workspace

```bash
# 列出所有Workspace
codepod list

# 启动Workspace
codepod start myproject

# 停止Workspace
codepod stop myproject

# 删除Workspace
codepod delete myproject
```

### 5. 连接Workspace

```bash
# SSH连接
ssh myproject.codepod

# 打开VS Code
codepod connect myproject --ide vscode
```

## 命令参考

| 命令 | 描述 |
|------|------|
| `codepod init` | 初始化配置 |
| `codepod config set <key> <value>` | 设置配置 |
| `codepod config list` | 查看配置 |
| `codepod up <name>` | 创建并启动Workspace (推荐) |
| `codepod create <name>` | 仅创建Workspace (不启动) |
| `codepod list` | 列出Workspace |
| `codepod start <name>` | 启动Workspace |
| `codepod stop <name>` | 停止Workspace |
| `codepod delete <name>` | 删除Workspace |
| `codepod connect <name>` | 连接Workspace |
| `codepod help` | 显示帮助 |

## 常见问题

### Q: WSL未安装
```
错误: WSL未安装
解决: 运行 wsl --install
```

### Q: Docker未运行
```
错误: Docker未运行
解决: 启动Docker Desktop
```

### Q: 端口被占用
```
错误: 端口2222被占用
解决: codepod config set general.ssh-port 2223
```

## 卸载

```bash
# 删除配置
rm -rf ~/.codepod

# 删除SSH配置
rm ~/.ssh/codepod*
```
