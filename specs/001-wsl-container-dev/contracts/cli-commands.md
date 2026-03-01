# CLI命令契约: codepod

## 1. init 命令

**用途**: 初始化配置文件

```bash
codepod init [--force]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| --force | bool | 否 | 强制初始化，覆盖现有配置 |

**输出**:
- 成功: 显示初始化成功信息
- 失败: 显示错误原因和解决方案

---

## 2. config 命令

### 2.1 config set

**用途**: 设置配置项

```bash
codepod config set <key> <value>
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| key | string | 是 | 配置键名 (如 wsl.distribution) |
| value | string | 是 | 配置值 |

**示例**:
```bash
codepod config set wsl.distribution Ubuntu-22.04
codepod config set wsl.docker-host tcp://localhost:2375
codepod config set general.ssh-port 2222
```

### 2.2 config list

**用途**: 列出所有配置

```bash
codepod config list [--format json|yaml]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| --format | string | 否 | 输出格式 (默认: yaml) |

### 2.3 config get

**用途**: 获取单个配置

```bash
codepod config get <key>
```

### 2.4 config reset

**用途**: 重置配置

```bash
codepod config reset [--all]
```

---

## 3. up 命令

**用途**: 创建并启动Workspace（类似 devpod up）

```bash
codepod up <name> [--repo <url>] [--ide <type>] [--path <dir>]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| name | string | 是 | Workspace名称 |
| --repo | string | 否 | Git仓库URL |
| --ide | string | 否 | IDE类型 (vscode/jetbrains) |
| --path | string | 否 | 本地代码目录 |

**示例**:
```bash
codepod up myproject --repo https://github.com/user/repo --ide vscode
codepod up myproject --path /home/user/myproject
```

---

## 4. create 命令

**用途**: 仅创建Workspace（不启动），类似 devpod create

```bash
codepod create <name> [--repo <url>] [--ide <type>] [--path <dir>]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| name | string | 是 | Workspace名称 |
| --repo | string | 否 | Git仓库URL |
| --ide | string | 否 | IDE类型 (vscode/jetbrains) |
| --path | string | 否 | 本地代码目录 |

**示例**:
```bash
codepod create myproject --repo https://github.com/user/repo
```

---

## 5. list 命令

**用途**: 列出所有Workspace

```bash
codepod list [--verbose]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| --verbose | bool | 否 | 显示详细信息 |

**输出格式**:
```
NAME        STATE      DOMAIN              CREATED
myproject   running    myproject.codepod   2026-03-01
```

---

## 5. start 命令

**用途**: 启动Workspace

```bash
codepod start <name>
```

---

## 6. stop 命令

**用途**: 停止Workspace

```bash
codepod stop <name>
```

---

## 7. delete 命令

**用途**: 删除Workspace

```bash
codepod delete <name> [--force]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| name | string | 是 | Workspace名称 |
| --force | bool | 否 | 强制删除 |

---

## 8. connect 命令

**用途**: 连接Workspace，自动启动IDE

```bash
codepod connect <name> [--ide <type>]
```

**参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| name | string | 是 | Workspace名称 |
| --ide | string | 否 | IDE类型 (默认: 配置中的IDE) |

**说明**:
- 连接后自动启动配置的IDE并连接到开发容器
- 支持的IDE: vscode, intellij-idea, goland, pycharm, webstorm

**示例**:
```bash
# 自动打开IDE并连接
codepod connect myproject

# 指定IDE类型
codepod connect myproject --ide vscode
codepod connect myproject --ide goland
```

---

## 9. up 命令 (含IDE启动)

**用途**: 创建并启动Workspace，自动打开IDE

```bash
codepod up <name> [--repo <url>] [--ide <type>] [--path <dir>]
```

**说明**:
- 创建Workspace后自动启动
- 启动后自动打开配置的IDE并连接到开发容器
- 等同于: codepod create + codepod start + codepod connect

**示例**:
```bash
# 创建并启动，自动打开VS Code
codepod up myproject --repo https://github.com/user/repo --ide vscode
```

---

## 10. help 命令

**用途**: 显示帮助

```bash
codepod help [<command>]
```

---

## 退出码

| 退出码 | 描述 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 2 | 参数错误 |
| 3 | 配置错误 |
| 4 | Workspace错误 |
| 5 | WSL错误 |
| 6 | Docker错误 |

---

## 错误格式

```json
{
  "error": {
    "code": 5,
    "message": "WSL未安装",
    "solution": "请运行 'wsl --install' 安装WSL2"
  }
}
```
