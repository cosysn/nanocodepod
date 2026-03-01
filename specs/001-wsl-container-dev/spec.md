# Feature Specification: WSL容器开发环境

**Feature Branch**: `001-wsl-container-dev`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "在windows上基于wsl构建一个容器开发环境。 类似devpod，daytona，coder这些，但是会简单很多。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - TUI界面 (Priority: P1)

系统需要提供一个漂亮的TUI（终端用户界面），在执行命令时显示友好的进度信息和执行过程，让用户清楚知道系统在做什么。

**Why this priority**: 良好的用户体验让命令行工具更易用，清晰的进度显示让用户知道系统正在工作。

**Independent Test**: 可以通过执行命令并验证TUI显示来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户首次使用或查看帮助，**When** 用户执行帮助命令，**Then** TUI显示友好的使用说明，包括可用命令、示例和基本操作指南
2. **Given** 用户执行命令，**When** 系统开始执行，**Then** TUI界面显示当前正在执行的操作（如"正在调用devcon命令制作镜像中..."）
3. **Given** devcon正在制作镜像，**When** 构建过程进行，**Then** TUI显示构建进度、当前步骤、日志输出
4. **Given** 命令执行成功，**Then** TUI显示成功信息和结果摘要
5. **Given** 命令执行失败，**Then** TUI显示错误原因（如"WSL未安装"、"Docker未运行"）
6. **Given** 命令执行失败，**Then** TUI显示解决方案（如"请运行 'wsl --install' 安装WSL"、"请启动Docker Desktop"）
7. **Given** 用户执行多个操作，**Then** TUI显示操作队列和当前进度

---

### User Story 2 - 配置存储 (Priority: P1)

系统需要将所有配置文件存放在用户目录的 .codepod 目录下，包括系统配置、Workspace信息等。

**Why this priority**: 配置文件是系统运行的基础，需要统一管理，便于迁移和备份。

**Independent Test**: 可以通过检查~/.codepod目录结构来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户首次使用系统，**When** 系统初始化，**Then** 自动创建~/.codepod目录
2. **Given** 用户创建Workspace，**When** 系统保存Workspace信息，**Then** 将信息保存到~/.codepod/workspaces/目录
3. **Given** 用户查看配置，**When** 执行查看配置命令，**Then** 显示~/.codepod目录下的配置文件内容

---

### User Story 3 - CLI配置管理 (Priority: P1)

用户需要通过命令行配置和管理系统参数，如指定使用哪个WSL的Docker、默认镜像仓库、SSH端口等。

**Why this priority**: 用户需要灵活配置系统行为，适应不同开发环境需求。

**Independent Test**: 可以通过执行配置命令并验证配置正确保存来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户需要指定WSL发行版，**When** 执行配置命令设置WSL名称，**Then** 系统保存配置并在后续操作中使用该WSL
2. **Given** 用户需要配置Docker端点，**When** 执行配置命令设置Docker host，**Then** 系统使用指定的Docker创建容器
3. **Given** 用户需要查看当前配置，**When** 执行查看配置命令，**Then** 显示所有当前配置项
4. **Given** 用户需要重置配置，**When** 执行重置配置命令，**Then** 恢复默认配置

---

### User Story 4 - 创建开发环境 (Priority: P1)

用户希望在Windows上快速创建一个基于WSL的容器化开发环境，只需一条命令即可启动完整的开发工作站。

**Why this priority**: 这是核心功能，用户使用该产品的首要目的就是创建开发环境。没有它其他功能毫无意义。

**Independent Test**: 可以通过执行"创建环境"命令并验证容器是否在WSL中成功启动来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户已安装WSL2和Docker Desktop，**When** 用户执行创建环境命令，**Then** 系统在WSL中创建并启动开发容器
2. **Given** 用户首次使用，**When** 执行创建命令，**Then** 系统自动配置必要的网络和存储卷
3. **Given** 用户已有开发环境，**When** 执行创建命令，**Then** 系统提示环境已存在或询问是否重建

---

### User Story 5 - 管理开发环境 (Priority: P1)

用户需要能够查看、启动、停止、删除已创建的开发环境。

**Why this priority**: 基本的生命周期管理功能，用户需要控制自己的开发环境资源。

**Independent Test**: 可以通过执行管理命令（列表/启动/停止/删除）并验证容器状态变化来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户已有开发环境，**When** 执行列出环境命令，**Then** 显示所有环境的名称、状态、创建时间
2. **Given** 环境处于停止状态，**When** 执行启动命令，**Then** 容器在WSL中重新启动
3. **Given** 环境处于运行状态，**When** 执行停止命令，**Then** 容器正常停止
4. **Given** 用户不再需要某环境，**When** 执行删除命令，**Then** 容器及其相关数据被清理

---

### User Story 6 - 连接开发环境 (Priority: P1)

用户需要能够通过SSH或端口转发连接到运行中的开发环境。

**Why this priority**: 用户创建开发环境后需要能够实际使用它进行编码和调试。

**Independent Test**: 可以通过执行连接命令并验证能够访问容器内的服务和shell来独立测试。

**Acceptance Scenarios**:

1. **Given** 环境处于运行状态，**When** 用户执行连接命令，**Then** 打开一个连接到容器内的Shell会话
2. **Given** 容器内运行着开发服务（如HTTP服务器），**When** 用户请求访问服务，**Then** 请求被正确转发到容器内的服务端口
3. **Given** 用户使用VS Code，**When** 用户请求连接，**Then** 系统提供Remote-WSL或SSH连接配置

---

### User Story 7 - 管理Workspace (Priority: P1)

用户需要能够管理Workspace，每个Workspace关联代码仓、IDE、开发容器、域名和SSH配置，实现完整的开发工作站管理。

**Why this priority**: Workspace是组织开发工作的核心单元，关联所有开发资源，提供一致的开发体验。

**Independent Test**: 可以通过创建和管理Workspace并验证所有关联资源正确配置来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户需要一个新项目的工作空间，**When** 创建Workspace并关联代码仓URL，**Then** 系统自动克隆代码到开发容器中
2. **Given** Workspace已创建，**When** 用户配置IDE偏好设置，**Then** 系统保存配置并在连接时自动应用
3. **Given** 用户需要通过域名访问开发服务，**When** 为Workspace配置域名，**Then** 系统自动配置DNS或hosts记录
4. **Given** Workspace创建成功，**When** 系统配置SSH config，**Then** 在SSH config中添加域名条目（格式：workspace_name.codepod）
5. **Given** Workspace创建成功，**When** 系统配置域名，**Then** 在/etc/hosts中添加域名映射（格式：workspace_name.codepod -> 容器IP）
6. **Given** 用户需要SSH连接到Workspace，**When** 使用域名连接，**Then** SSH config正确解析域名并连接到Agent
7. **Given** Workspace包含代码仓、IDE配置、容器、域名、SSH配置，**When** 用户列出Workspace，**Then** 显示所有关联资源的摘要信息

---

### User Story 8 - Agent注入 (Priority: P1)

用户需要在开发容器中注入一个Agent，该Agent提供SSH Server（不需要安装sshd）、Git Forward和监控功能。

**Why this priority**: Agent是连接宿主机和容器的核心组件，提供轻量级的远程访问能力，无需传统SSH服务。

**Independent Test**: 可以通过启动容器并验证Agent服务正常运行来独立测试。

**Acceptance Scenarios**:

1. **Given** 开发容器已启动，**When** 系统注入Agent，**Then** Agent在容器内自动运行，提供SSH Server功能
2. **Given** Agent已运行，**When** 用户通过SSH连接Agent，**Then** 可以访问容器内的Shell（无需安装sshd）
3. **Given** Agent已运行，**When** 用户执行Git操作，**Then** Agent提供Git代理转发，确保Git命令正常工作
4. **Given** Agent已运行，**When** 用户查询容器状态，**Then** Agent提供CPU、内存、网络等监控数据
5. **Given** Agent故障，**When** 系统检测到Agent异常，**Then** Agent自动重启或通知用户

---

### User Story 9 - Devcontainer支持 (Priority: P1)

系统需要支持将 devcon 工具注入到WSL，利用该工具基于 .devcontainer 规范制作开发容器镜像，并使用Docker创建开发容器。

**Why this priority**: 兼容devcontainer规范可以利用现有开发环境配置，实现标准化开发环境构建。

**Independent Test**: 可以通过指定代码目录包含.devcontainer.json并验证容器镜像正确构建来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户有devcon工具，**When** 系统初始化，**Then** 将devcon工具注入到WSL中
2. **Given** 代码目录包含.devcontainer.json，**When** 用户请求创建开发容器，**Then** devcon工具根据配置构建开发容器镜像
3. **Given** 开发容器镜像已构建，**When** 系统创建开发容器，**Then** 使用Docker在WSL中运行开发容器
4. **Given** 代码目录包含自定义Dockerfile，**Then** devcon工具支持使用自定义Dockerfile构建镜像

---

### User Story 10 - 自定义开发环境 (Priority: P2)

用户需要能够自定义开发环境的配置，如选择编程语言、工具版本、资源限制等。

**Why this priority**: 不同项目需要不同的开发环境配置，用户需要灵活性。

**Independent Test**: 可以通过指定自定义配置创建环境并验证环境包含所需工具来独立测试。

**Acceptance Scenarios**:

1. **Given** 用户指定需要的编程语言和版本，**When** 创建环境，**Then** 容器内包含指定的语言运行时
2. **Given** 用户指定资源限制（CPU/内存），**Then** 容器按照指定限制运行
3. **Given** 用户提供自定义Dockerfile，**Then** 系统使用该Dockerfile构建环境

---

### Edge Cases

- 当WSL未安装或版本过低时，系统必须提示用户安装WSL2
- 当Docker Desktop未运行或磁盘空间不足时，必须给出明确错误提示
- 当网络连接中断时，环境操作应该能够安全重试
- 当容器内部进程崩溃时，应该能够自动重启或通知用户

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 提供漂亮的TUI（终端用户界面）
- **FR-002**: 系统 MUST 在执行命令时显示友好的进度信息
- **FR-003**: 系统 MUST 显示当前正在执行的操作（如"正在调用devcon命令制作镜像中..."）
- **FR-004**: 系统 MUST 显示命令执行过程中的日志输出
- **FR-005**: 系统 MUST 在命令执行成功时显示成功信息和结果摘要
- **FR-005a**: 系统 MUST 在用户首次使用或查看帮助时显示友好的使用说明
- **FR-005b**: 系统 MUST 在命令执行失败时显示错误原因（如WSL未安装、Docker未运行）
- **FR-005c**: 系统 MUST 在命令执行失败时显示解决方案（如请运行XX命令）
- **FR-006**: 系统 MUST 在命令执行失败时显示错误信息和解决方案
- **FR-007**: 系统 MUST 将所有配置文件存放在用户目录的 ~/.codepod 目录下
- **FR-008**: 系统 MUST 在 ~/.codepod/workspaces/ 目录下保存Workspace信息
- **FR-009**: 系统 MUST 支持配置文件格式（JSON/YAML）
- **FR-010**: 系统 MUST 支持命令行配置管理（如设置使用哪个WSL的Docker）
- **FR-011**: 系统 MUST 支持配置WSL发行版名称
- **FR-012**: 系统 MUST 支持配置Docker端点
- **FR-013**: 系统 MUST 支持查看当前配置
- **FR-014**: 系统 MUST 支持重置配置到默认值
- **FR-015**: 系统 MUST 提供命令行工具用于环境管理
- **FR-016**: 系统 MUST 支持在WSL2中创建和管理Docker容器
- **FR-017**: 系统 MUST 支持将devcon工具注入到WSL
- **FR-018**: 系统 MUST 支持基于.devcontainer.json配置构建开发容器镜像
- **FR-019**: 系统 MUST 支持使用Docker创建开发容器
- **FR-020**: 系统 MUST 提供环境创建、列表、启动、停止、删除命令
- **FR-021**: 系统 MUST 支持SSH连接到运行中的容器
- **FR-022**: 系统 MUST 支持端口转发使容器内服务可从Windows访问
- **FR-023**: 系统 MUST 支持通过配置文件自定义开发环境（编程语言、工具、资源限制）
- **FR-024**: 系统 MUST 验证WSL2和Docker Desktop已正确安装和运行
- **FR-025**: 系统 MUST 提供清晰的错误信息和故障排除指引
- **FR-026**: 系统 MUST 支持创建和管理Workspace
- **FR-027**: 系统 MUST 支持Workspace关联代码仓（Git仓库URL）
- **FR-028**: 系统 MUST 支持Workspace关联IDE配置（VS Code、JetBrains等）
- **FR-029**: 系统 MUST 支持Workspace关联开发容器
- **FR-030**: 系统 MUST 支持Workspace配置访问域名（格式：workspace_name.codepod）
- **FR-031**: 系统 MUST 支持在SSH config中配置域名条目
- **FR-032**: 系统 MUST 支持在/etc/hosts中配置域名映射
- **FR-033**: 系统 MUST 支持在开发容器中注入Agent
- **FR-034**: Agent MUST 提供SSH Server功能（不依赖sshd）
- **FR-035**: Agent MUST 提供Git Forward代理功能
- **FR-036**: Agent MUST 提供容器监控功能（CPU、内存、网络）

### Key Entities

- **配置目录**: 系统配置文件存放目录 ~/.codepod，包含workspaces、config等子目录
- **Workspace**: 代表一个完整的开发工作空间，关联代码仓、IDE、开发容器、域名、SSH配置
- **代码仓**: Git仓库的URL和认证信息
- **IDE配置**: 用户偏好的IDE设置和插件
- **开发环境**: 代表一个完整的开发工作站，包含容器、配置、状态
- **环境配置**: 定义环境的属性（基础镜像、资源限制、预装工具、端口映射）
- **环境状态**: 描述环境的当前状态（未创建、运行中、已停止、错误）
- **域名配置**: Workspace的访问域名和DNS配置
- **SSH配置**: 连接Workspace所需的SSH密钥和配置
- **Agent**: 注入到容器的轻量级服务进程，提供SSH Server、Git Forward、监控功能

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 用户在5分钟内完成从安装到创建第一个开发环境
- **SC-002**: 开发环境启动时间不超过60秒
- **SC-003**: 100%的用户能够成功创建基础开发环境
- **SC-004**: 用户满意度评分达到4分以上（5分制）
- **SC-005**: 内存和CPU使用率符合配置的限制值
