# Claude Code (Go 实现版本)

> Claude Code 的 Go 语言开源实现, Anthropic 的 AI 编程助手

**[English](README.md) | 中文版**

## 概述

Claude Code Go 是 [Claude Code](https://github.com/anthropics/claude-code) 的开源 Go 语言重实现版本,提供以下核心功能:

- **交互式 REPL 模式** - 在终端中与 Claude 对话
- **终端用户界面 (TUI)** - 基于 Bubble Tea 的富文本界面
- **工具系统** - 执行命令、读写文件、搜索等
- **斜杠命令系统** - `/commit`、`/review`、`/help` 以及 40+ 其他命令
- **MCP 协议支持** - 连接 Model Context Protocol 服务器
- **插件和技能系统** - 通过插件和技能扩展功能

## 项目架构

```
claude-code-go/
├── cmd/claude/           # CLI 入口点
├── internal/
│   ├── cli/             # CLI 命令 (基于 Cobra)
│   ├── engine/          # 核心查询引擎
│   │   ├── query_engine.go   # 主查询循环
│   │   ├── tool.go           # 工具接口和注册表
│   │   ├── command.go        # 命令接口和注册表
│   │   ├── context.go        # 上下文管理
│   │   └── permission.go      # 权限系统
│   ├── tools/           # 工具实现
│   │   ├── builtins.go       # Bash、Read、Write、Glob、Grep
│   │   ├── web.go           # WebFetch、WebSearch、Todo、Edit
│   │   ├── task.go          # 任务工具
│   │   ├── agent.go         # Agent 工具
│   │   ├── mcp.go           # MCP 工具
│   │   └── additional.go     # 附加工具 (Plan、Worktree 等)
│   ├── commands/        # 斜杠命令实现
│   ├── services/        # API 客户端和服务
│   │   ├── analytics/      # 分析和特性开关
│   │   ├── compact/        # 上下文压缩
│   │   ├── lsp/           # 语言服务器协议
│   │   ├── mcp/           # MCP 协议客户端
│   │   ├── oauth/         # OAuth 2.0 认证
│   │   ├── plugins/       # 插件系统
│   │   ├── settings/       # 设置管理
│   │   ├── skills/        # 技能管理
│   │   ├── team/          # 团队记忆同步
│   │   ├── token/         # 令牌估算
│   │   └── voice/         # 语音输入输出
│   ├── ui/              # Bubble Tea TUI
│   ├── bridge/          # IDE 桥接 (VS Code、JetBrains)
│   ├── protocol/        # JSON-RPC 协议
│   └── storage/         # 会话和配置存储
└── pkg/
    └── anthropic/       # Anthropic API 客户端
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 编程语言 | Go 1.21+ |
| CLI 框架 | [Cobra](https://github.com/spf13/cobra) |
| TUI 框架 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + Lipgloss |
| HTTP 客户端 | net/http (标准库) |
| 配置管理 | [Viper](https://github.com/spf13/viper) |
| 日志 | zap |

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- Anthropic API 密钥

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/claude-code-go.git
cd claude-code-go

# 下载依赖
go mod download

# 构建 CLI
go build -o claude ./cmd/claude
```

### 配置

```bash
# 通过环境变量设置 API 密钥
export ANTHROPIC_API_KEY="your-api-key"

# 或者使用登录命令
./claude login
```

## 使用方法

### 交互模式

```bash
# 启动交互式 REPL 模式
./claude repl

# 启动终端 UI 模式
./claude tui
```

### 单次任务

```bash
# 执行单个任务
./claude "写一个 Go 的 Hello World 程序"
```

### 斜杠命令

斜杠命令以 `/` 为前缀,提供专业功能:

#### 核心命令
| 命令 | 描述 |
|------|------|
| `/help` | 显示帮助信息 |
| `/model [名称]` | 获取或设置 AI 模型 |
| `/config get/set/list` | 管理配置 |
| `/cost` | 显示会话使用量和费用 |
| `/exit` | 退出 Claude Code |

#### Git 命令
| 命令 | 描述 |
|------|------|
| `/commit` | 创建 Git 提交 |
| `/diff` | 显示变更 |
| `/branch` | 管理分支 |
| `/logs` | 查看 Git 日志 |

#### 代码命令
| 命令 | 描述 |
|------|------|
| `/review` | 审查代码变更 |
| `/ultrareview` | 综合代码审查 |
| `/compact` | 压缩上下文以节省令牌 |

#### 开发命令
| 命令 | 描述 |
|------|------|
| `/init` | 初始化项目 |
| `/mcp` | 管理 MCP 服务器 |
| `/skills` | 管理技能 |
| `/doctor` | 运行诊断检查 |

### CLI 参数

```bash
-v, --verbose         # 详细输出
--json              # JSON 格式输出
--config <路径>     # 配置文件路径
```

## 工具列表

Claude Code 提供以下工具与系统交互:

### 内置工具
| 工具 | 权限 | 描述 |
|------|------|------|
| `Bash` | 提升权限 | 执行 shell 命令 |
| `Read` | 只读 | 读取文件内容 |
| `Write` | 提升权限 | 写入文件内容 |
| `Glob` | 只读 | 按模式查找文件 |
| `Grep` | 只读 | 搜索文件内容 |

### Web 工具
| 工具 | 权限 | 描述 |
|------|------|------|
| `WebFetch` | 只读 | 获取 URL 内容 |
| `WebSearch` | 只读 | 网络搜索 |
| `TodoWrite` | 普通 | 管理待办事项 |
| `Edit` | 提升权限 | 编辑文件 (替换文本) |

### 任务工具
| 工具 | 权限 | 描述 |
|------|------|------|
| `TaskCreate` | 普通 | 创建任务 |
| `TaskList` | 只读 | 列出所有任务 |
| `TaskUpdate` | 普通 | 更新任务状态 |
| `TaskGet` | 只读 | 获取任务详情 |
| `TaskStop` | 普通 | 停止任务 |

### 附加工具
| 工具 | 权限 | 描述 |
|------|------|------|
| `AskUserQuestion` | 普通 | 向用户提问 |
| `EnterPlanMode` | 普通 | 进入计划模式 |
| `ExitPlanMode` | 普通 | 退出计划模式 |
| `EnterWorktree` | 提升权限 | 进入 Git worktree |
| `ExitWorktree` | 提升权限 | 退出 Git worktree |
| `NotebookEdit` | 提升权限 | 编辑 Jupyter 笔记本 |
| `Skill` | 普通 | 执行技能 |
| `LSP` | 只读 | LSP 操作 |
| `McpAuth` | 提升权限 | MCP 认证 |
| `PowerShell` | 提升权限 | 执行 PowerShell 命令 |

### Agent 工具
| 工具 | 权限 | 描述 |
|------|------|------|
| `Agent` | 普通 | 创建子 Agent |
| `AgentResult` | 普通 | 获取 Agent 结果 |
| `SendMessage` | 普通 | 发送消息 |
| `TeamCreate` | 普通 | 创建 Agent 团队 |
| `TeamDelete` | 普通 | 删除团队 |

## 服务模块

项目包含以下服务:

| 服务 | 描述 |
|------|------|
| `analytics` | 事件追踪和特性开关 |
| `compact` | 长对话的上下文压缩 |
| `lsp` | 语言服务器协议集成 |
| `mcp` | Model Context Protocol 客户端 |
| `oauth` | OAuth 2.0 认证 |
| `plugins` | 插件系统扩展 |
| `settings` | 配置管理 |
| `skills` | 技能管理系统 |
| `team` | 团队记忆同步 |
| `token` | 令牌估算和成本计算 |
| `voice` | 语音输入输出支持 |

## 测试

```bash
# 运行所有测试
go test ./...

# 详细输出
go test -v ./...

# 带覆盖率
go test -cover ./...

# 运行特定包的测试
go test -v ./internal/services/token/...
```

## 项目状态

### 已实现
- [x] 基于 Cobra 框架的 CLI
- [x] 工具接口和注册表
- [x] 命令接口和注册表
- [x] Anthropic API 客户端
- [x] 40+ 内置工具
- [x] 带工具执行循环的查询引擎
- [x] 权限系统
- [x] 上下文管理
- [x] Bubble Tea TUI
- [x] 11 个服务模块及测试
- [x] 全面的测试覆盖

### 开发中
- [ ] 完整的斜杠命令套件
- [ ] 完整的 MCP 协议支持
- [ ] 上下文压缩实现
- [ ] 插件系统
- [ ] 技能系统
- [ ] IDE 桥接集成

## 贡献

欢迎贡献! 请:

1. 阅读架构文档
2. 遵循现有代码风格
3. 为新功能添加测试
4. 提交前运行 `go vet` 和 `gofmt`

## 许可证

本项目仅供教育目的。原版 Claude Code 版权归 Anthropic 所有。

## 参考资料

- [Claude Code 文档](https://docs.anthropic.com/claude-code)
- [Anthropic API](https://docs.anthropic.com)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Cobra](https://github.com/spf13/cobra)
