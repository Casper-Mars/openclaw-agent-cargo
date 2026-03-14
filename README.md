# OpenClaw Agent Cargo

OpenClaw Agent 导入/导出工具 - 用于备份和恢复 OpenClaw Agent 配置。

## 功能

- ✅ 导出单个 Agent 到 tar.gz 包
- ✅ 从 tar.gz 包导入 Agent
- ✅ 支持自定义导出路径
- ✅ 支持导入时重命名 Agent

## 安装

### 从源码编译

```bash
git clone https://github.com/Casper-Mars/openclaw-agent-cargo.git
cd openclaw-agent-cargo
go build -o openclaw-agent-cargo
```

### 使用 Go Install

```bash
go install github.com/Casper-Mars/openclaw-agent-cargo@latest
```

## 使用方法

### 导出 Agent

```bash
# 导出到当前目录
./openclaw-agent-cargo export --agent main

# 导出到指定路径
./openclaw-agent-cargo export --agent main --output ~/backups/main-agent.tar.gz

# 导出时包含详细日志
./openclaw-agent-cargo export --agent main --verbose
```

### 导入 Agent

```bash
# 从当前目录导入
./openclaw-agent-cargo import --file main-agent.tar.gz

# 从指定路径导入
./openclaw-agent-cargo import --file ~/backups/main-agent.tar.gz

# 导入时覆盖现有 Agent
./openclaw-agent-cargo import --file main-agent.tar.gz --force

# 导入时重命名 Agent
./openclaw-agent-cargo import --file main-agent.tar.gz --rename backup-main
```

## 导出包结构

```
agent-export.tar.gz
├── manifest.json           # 包元信息
├── config.json             # Agent 配置
├── sessions/
│   └── sessions.json       # 会话存储
├── workspace/              # 工作区文件
│   ├── AGENTS.md
│   ├── SOUL.md
│   ├── USER.md
│   ├── IDENTITY.md
│   ├── TOOLS.md
│   ├── HEARTBEAT.md
│   ├── MEMORY.md
│   └── memory/
└── transcripts/            # 会话转录（可选）
```

## 系统要求

- Go 1.21+
- OpenClaw 安装目录：`~/.openclaw/`

## 许可证

MIT License
